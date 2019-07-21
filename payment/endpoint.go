package payment

import (
	"context"

	"github.com/otetz/payments/account"

	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
)

type errorOnlyResponse struct {
	Err error `json:"error,omitempty"`
}

func (r errorOnlyResponse) ErrError() error { return r.Err }

type newPaymentRequest struct {
	FromAccountID account.ID      `json:"from" valid:"alphanum,required"`
	Amount        decimal.Decimal `json:"amount" valid:"decimal,required"`
	ToAccountID   account.ID      `json:"to" valid:"alphanum,required"`
}

func makeNewPaymentEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(newPaymentRequest)
		err := s.New(req.FromAccountID, req.Amount, req.ToAccountID)
		return errorOnlyResponse{Err: err}, nil
	}
}

type loadPaymentsRequest struct {
	AccountID account.ID `json:"account"`
}

func makeLoadPaymentsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loadPaymentsRequest)
		r := s.Load(req.AccountID)
		return r, nil
	}
}

func makeLoadAllPaymentsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := s.LoadAll()
		return r, nil
	}
}
