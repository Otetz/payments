package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-pg/pg"
	"github.com/otetz/payments/db"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/payment"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type dbLogger struct{}

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {
}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	fmt.Println(q.FormattedQuery())
}

func main() {
	httpAddr := ":8095"

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	conn := pg.Connect(&pg.Options{
		Addr:            "localhost:5432", // TODO: enviflag
		User:            "postgres",       // TODO: enviflag
		Password:        "example",        // TODO: enviflag
		Database:        "payments",       // TODO: enviflag
		ApplicationName: "payments",       // TODO: enviflag
		PoolSize:        10,               // TODO: enviflag
	})
	defer conn.Close()
	conn.AddQueryHook(dbLogger{})

	if err := db.CreateSchema(conn); err != nil {
		_ = logger.Log("transport", "DB", "address", "172.18.0.2", "msg", err)
	}

	//var (
	//	accounts = inmem.NewAccountRepository()
	//	payments = inmem.NewPaymentRepository(accounts)
	//)
	var (
		accounts = db.NewAccountRepository(conn)
		payments = db.NewPaymentRepository(conn, accounts)
	)

	fieldKeys := []string{"method"}

	as := account.NewService(accounts)
	as = account.NewLoggingService(log.With(logger, "component", "account"), as)
	as = account.NewMetricsService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "account_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "account_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		as,
	)

	ps := payment.NewService(payments, accounts)
	ps = payment.NewLoggingService(log.With(logger, "component", "payment"), ps)
	ps = payment.NewMetricsService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "paymemt_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "paymemt_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		ps,
	)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/api/accounts/v1/", account.MakeHandler(as, httpLogger))
	mux.Handle("/api/payments/v1/", payment.MakeHandler(ps, httpLogger))

	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error, 2)
	go func() {
		_ = logger.Log("transport", "http", "address", httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(httpAddr, nil)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	_ = logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
