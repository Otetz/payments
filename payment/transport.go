package payment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/errs"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// MakeHandler returns a handler for the payment service.
func MakeHandler(s Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(errs.EncodeError),
	}

	newPaymentHandler := kithttp.NewServer(
		makeNewPaymentEndpoint(s),
		decodeNewPaymentRequest,
		errs.EncodeResponse,
		opts...,
	)

	loadPaymentsHandler := kithttp.NewServer(
		makeLoadPaymentsEndpoint(s),
		decodeLoadPaymentsRequest,
		errs.EncodeResponse,
		opts...,
	)

	loadAllPaymentsHandler := kithttp.NewServer(
		makeLoadAllPaymentsEndpoint(s),
		decodeLoadAllPaymentsRequest,
		errs.EncodeResponse,
		opts...,
	)

	router := mux.NewRouter()

	router.Handle("/api/payments/v1/payments", newPaymentHandler).Methods("POST")
	router.Handle("/api/payments/v1/payments", loadAllPaymentsHandler).Methods("GET")
	router.Handle("/api/payments/v1/payments/{id}", loadPaymentsHandler).Methods("GET")

	return router
}

func decodeNewPaymentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body newPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	if _, err := govalidator.ValidateStruct(body); err != nil {
		return nil, errs.ValidationError{Err: err}
	}
	return body, nil
}

func decodeLoadPaymentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errs.ErrBadRoute
	}
	return loadPaymentsRequest{AccountID: account.ID(id)}, nil
}

func decodeLoadAllPaymentsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}
