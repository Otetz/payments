package account

import (
	"context"

	"github.com/otetz/payments/errs"

	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"
)

type idField struct {
	ID ID `json:"id" valid:"alphanum,required"`
}

type newAccountRequest struct {
	ID       ID              `json:"id" valid:"alphanum,required"`
	Currency Currency        `json:"currency" valid:"in(CurrencyUSD)"`
	Balance  decimal.Decimal `json:"balance,omitempty" valid:"decimal"`
}

func makeNewAccountEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(newAccountRequest)
		err := s.New(req.ID, req.Currency, req.Balance)
		return errs.ErrorOnlyResponse{Err: err}, nil
	}
}

type loadAccountResponse struct {
	Account *Account `json:"account,omitempty"`
	Err     error    `json:"error,omitempty"`
}

func (r loadAccountResponse) ErrError() error { return r.Err }

func makeLoadAccountEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(idField)
		a, err := s.Load(req.ID)
		return loadAccountResponse{Account: a, Err: err}, nil
	}
}

func makeLoadAllAccountsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := s.LoadAll()
		return r, nil
	}
}

func makeDeleteAccountEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(idField)
		err := s.Delete(req.ID)
		return errs.ErrorOnlyResponse{Err: err}, nil
	}
}
