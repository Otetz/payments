package main

import (
	"flag"
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

var (
	flagHttpAddr = flag.String("http_address", "0.0.0.0:8080", "Http address for web server running")

	flagDBAddr     = flag.String("db_address", "localhost:5432", "Address to connect to PostgreSQL server")
	flagDBUser     = flag.String("db_user", "postgres", "PostgreSQL connection user")
	flagDBPassword = flag.String("db_password", "", "PostgreSQL connection password")
	flagDBDatabase = flag.String("database", "payments", "PostgreSQL database name")
	flagDBAppName  = flag.String("app_name", "payments", "PostgreSQL application name (for logging)")
	flagDBPoolSize = flag.Int("pool_size", 10, "PostgreSQL connection pool size")
	flagDBLog = flag.Bool("db_log", false, "Switch for statements logging")
)

func main() {
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	conn := setupDB(logger)
	defer func() {
		if err := conn.Close(); err != nil {
			_ = logger.Log("error", err)
		}
	}()

	var (
		accounts = db.NewAccountRepository(conn)
		payments = db.NewPaymentRepository(conn, accounts)
	)

	as := setupAccountService(accounts, logger)
	ps := setupPaymentService(payments, accounts, logger)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/api/accounts/v1/", account.MakeHandler(as, httpLogger))
	mux.Handle("/api/payments/v1/", payment.MakeHandler(ps, httpLogger))

	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error, 2)
	go func() {
		_ = logger.Log("transport", "http", "address", *flagHttpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*flagHttpAddr, nil)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	_ = logger.Log("terminated", <-errs)
}

func setupDB(logger log.Logger) *pg.DB {
	conn := pg.Connect(&pg.Options{
		Addr:            *flagDBAddr,
		User:            *flagDBUser,
		Password:        *flagDBPassword,
		Database:        *flagDBDatabase,
		ApplicationName: *flagDBAppName,
		PoolSize:        *flagDBPoolSize,
	})
	if *flagDBLog {
		conn.AddQueryHook(dbLogger{})
	}
	if err := db.CreateSchema(conn); err != nil {
		_ = logger.Log("transport", "DB", "address", *flagDBAddr, "msg", err)
	}
	return conn
}

func setupPaymentService(payments payment.Repository, accounts account.Repository, logger log.Logger) payment.Service {
	fieldKeys := []string{"method"}

	ps := payment.NewService(payments, accounts)
	ps = payment.NewLoggingService(log.With(logger, "component", "payment"), ps)
	ps = payment.NewMetricsService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "payment_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "payment_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		ps,
	)
	return ps
}

func setupAccountService(accounts account.Repository, logger log.Logger) account.Service {
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
	return as
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
