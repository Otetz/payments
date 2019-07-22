package payment

import (
	"time"

	"github.com/otetz/payments/account"

	"github.com/go-kit/kit/metrics"
	"github.com/shopspring/decimal"
)

type metricsService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Service
}

// NewMetricsService returns an instance of a metrics Service.
func NewMetricsService(counter metrics.Counter, latency metrics.Histogram, s Service) Service {
	return &metricsService{
		requestCount:   counter,
		requestLatency: latency,
		Service:        s,
	}
}

// New is logging wrapper for new payment creation.
func (s *metricsService) New(fromAccountID account.ID, amount decimal.Decimal, toAccountID account.ID) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "new").Add(1)
		s.requestLatency.With("method", "new").Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	return s.Service.New(fromAccountID, amount, toAccountID)
}

// Load is logging wrapper for load payments by account.
func (s *metricsService) Load(accountID account.ID) []*Payment {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	return s.Service.Load(accountID)
}

// LoadAll is logging wrapper for load all payments.
func (s *metricsService) LoadAll() []*Payment {
	defer func(begin time.Time) {
		s.requestCount.With("method", "loadAll").Add(1)
		s.requestLatency.With("method", "loadAll").Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	return s.Service.LoadAll()
}
