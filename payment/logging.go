package payment

import (
	"time"

	"github.com/otetz/payments/account"

	"github.com/go-kit/kit/log"
	"github.com/shopspring/decimal"
)

type loggingService struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

// New is logging wrapper for new payment creation.
func (s *loggingService) New(fromAccountID account.ID, amount decimal.Decimal, toAccountID account.ID) (err error) {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			"method", "new",
			"from", fromAccountID,
			"amount", amount,
			"to", toAccountID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.New(fromAccountID, amount, toAccountID)
}

// Load is logging wrapper for load payments by account.
func (s *loggingService) Load(accountID account.ID) (result []*Payment) {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			"method", "loadForAccount",
			"account_id", accountID,
			"len(result)", len(result),
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.Service.Load(accountID)
}

// LoadAll is logging wrapper for load all payments.
func (s *loggingService) LoadAll() (result []*Payment) {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			"method", "loadForAccount",
			"len(result)", len(result),
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.Service.LoadAll()
}
