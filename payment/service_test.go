package payment_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/otetz/payments/payment"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/inmem"
	"github.com/shopspring/decimal"
)

func OK(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

type Case struct {
	Name             string
	Method           string
	Path             string
	Payload          interface{}
	PayloadParameter PayloadParameter
	Status           int
	Result           interface{}
	CheckRepo        bool
}

type PayloadParameter struct {
	Name   string
	Values []string
}

type CaseRequestPayload map[string]interface{}
type CaseResponse map[string]interface{}

const (
	EndpointURL = "/api/payments/v1/payments"
)

func TestPaymentApi(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpLogger := log.With(logger, "component", "http")

	accounts := inmem.NewAccountRepository()
	payments := inmem.NewPaymentRepository(accounts)
	ps := payment.NewService(payments, accounts)

	handler := payment.MakeHandler(ps, httpLogger)

	_ = accounts.Store(&account.Account{ID: "test1", Balance: decimal.NewFromFloat(1000.0), Currency: "USD"})
	_ = accounts.Store(&account.Account{ID: "test2", Currency: "USD"})

	_ = payments.Store(&payment.Payment{
		ID:        uuid.New(),
		Account:   "test1",
		Amount:    decimal.NewFromFloat(55.55),
		ToAccount: "test2",
		Direction: payment.Outgoing,
	})
	_ = payments.Store(&payment.Payment{
		ID:          uuid.New(),
		Account:     "test2",
		Amount:      decimal.NewFromFloat(55.55),
		FromAccount: "test1",
		Direction:   payment.Incoming,
	})

	cases := []Case{
		{
			Name:   "load for account:normal flow",
			Path:   EndpointURL + "/test1",
			Method: http.MethodGet,
			Status: http.StatusOK,
			Result: []CaseResponse{
				{
					"account":    "test1",
					"amount":     55.55,
					"to_account": "test2",
					"direction":  "outgoing",
				},
			},
		},
		{
			Name:   "load all payments:normal flow",
			Path:   EndpointURL,
			Method: http.MethodGet,
			Status: http.StatusOK,
			Result: []CaseResponse{
				{
					"account":    "test1",
					"amount":     55.55,
					"to_account": "test2",
					"direction":  "outgoing",
				},
				{
					"account":      "test2",
					"amount":       55.55,
					"from_account": "test1",
					"direction":    "incoming",
				},
			},
		},
		{
			Name:   "new payment:normal flow",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test2"),
			},
			Status: http.StatusOK,
			Result: CaseResponse{},
		},
		{
			Name:   "new payment:incorrect decimal",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1"),
				"amount": "33,33",
				"to":     account.ID("test2"),
			},
			Status: http.StatusInternalServerError,
			Result: CaseResponse{"error": "Error decoding string '33,33': can't convert 33,33 to decimal"},
		},
		{
			Name:   "new payment:incorrect source account",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID(strings.Repeat("abcd1234", 33)),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test2"),
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "validation error: from: " + strings.Repeat("abcd1234", 33) +
				" does not validate as stringlength(1|255)"},
		},
		{
			Name:   "new payment:incorrect target account",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test2"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID(strings.Repeat("abcd1234", 33)),
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "validation error: to: " + strings.Repeat("abcd1234", 33) +
				" does not validate as stringlength(1|255)"},
		},
		{
			Name:   "new payment:validate account",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1фыва"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test2"),
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "validation error: from: test1фыва does not validate as alphanum"},
		},
		{
			Name:   "new payment:wrong source account",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test3"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test2"),
			},
			Status: http.StatusNotFound,
			Result: CaseResponse{"error": "unknown source account"},
		},
		{
			Name:   "new payment:wrong target account",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test3"),
			},
			Status: http.StatusNotFound,
			Result: CaseResponse{"error": "unknown target account"},
		},
		{
			Name:   "new payment:accounts the same",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1"),
				"amount": decimal.NewFromFloat(33.33),
				"to":     account.ID("test1"),
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "target account must not be equal to source account"},
		},
		{
			Name:   "new payment:insufficient money",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"from":   account.ID("test1"),
				"amount": decimal.NewFromFloat(9999),
				"to":     account.ID("test2"),
			},
			Status: http.StatusBadRequest,
			Result: CaseResponse{"error": "insufficient money on source account"},
		},
	}

	runTests(t, handler, cases, accounts)

	t.Run("new payment:wrong json", func(t *testing.T) {
		payload := `{ "a":1 `

		req, err := http.NewRequest("POST", EndpointURL, strings.NewReader(payload))
		OK(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", status,
				http.StatusInternalServerError)
		}
		expectedBody := `{"error":"unexpected EOF"}`
		if strings.TrimSpace(rr.Body.String()) != expectedBody {
			t.Errorf("handler returned wrong body: got %v want %v", rr.Body.String(), expectedBody)
		}
	})
}

func runTests(t *testing.T, handler http.Handler, cases []Case, repository account.Repository) {
	for idx, item := range cases {
		idx := idx
		item := item
		var (
			err      error
			req      *http.Request
			result   interface{}
			expected interface{}
		)

		if item.Name == "" {
			item.Name = fmt.Sprintf("[%s] %s", item.Method, item.Path)
		}
		caseName := fmt.Sprintf("[%d]:%s", idx, item.Name)

		t.Run(caseName, func(t *testing.T) {
			payloads := make([]string, 0)

			if item.PayloadParameter.Name == "" {
				payload, err := json.Marshal(&item.Payload)
				OK(t, err)
				payloads = append(payloads, string(payload))
			} else {
				for _, val := range item.PayloadParameter.Values {
					payloadStruct := item.Payload.(CaseRequestPayload)
					payloadStruct[item.PayloadParameter.Name] = val
					payload, err := json.Marshal(&payloadStruct)
					OK(t, err)
					payloads = append(payloads, string(payload))
				}
			}

			for _, payload := range payloads {
				req, err = http.NewRequest(item.Method, item.Path, strings.NewReader(payload))
				OK(t, err)

				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if status := rr.Code; status != item.Status {
					t.Errorf("[%s] handler returned wrong status code: got %v want %v",
						caseName, status, item.Status)
				}

				if item.Result != nil {
					body, err := ioutil.ReadAll(rr.Body)
					OK(t, err)
					err = json.Unmarshal(body, &result)
					OK(t, err)

					expectedBody, err := json.Marshal(&item.Result)
					OK(t, err)
					expectedBodyStr := string(expectedBody)
					if strings.Contains(expectedBodyStr, "*") {
						var validBody = regexp.MustCompile(`^` + expectedBodyStr + `$`)
						if !validBody.MatchString(strings.TrimSpace(string(body))) {
							t.Errorf("[%s] handler returned wrong body: got %v want %v", caseName, string(body),
								expectedBodyStr)
						}
					} else {
						_ = json.Unmarshal(expectedBody, &expected)

						if !reflect.DeepEqual(result, expected) {
							t.Errorf("[%d] results not match\nGot: %#v\nExpected: %#v", idx, result, expected)
							continue
						}
					}
				}

				if item.CheckRepo {
					payloadStruct := item.Payload.(CaseRequestPayload)
					a, _ := repository.Find(payloadStruct["id"].(account.ID))
					expectedAcc := &account.Account{
						ID:       payloadStruct["id"].(account.ID),
						Balance:  decimal.NewFromFloat(0.0),
						Currency: account.CurrencyUSD,
						Deleted:  false,
					}
					if payloadStruct["balance"] != nil {
						expectedAcc.Balance = payloadStruct["balance"].(decimal.Decimal)
					}
					if payloadStruct["currency"] != nil {
						expectedAcc.Currency = payloadStruct["currency"].(account.Currency)
					}
					if !cmp.Equal(a, expectedAcc) {
						t.Errorf("[%s] store returned not equal account: got %v want %v", caseName, a, expectedAcc)
					}
				}
			}
		})
	}
}
