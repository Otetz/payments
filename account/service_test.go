package account_test

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
	EndpointURL = "/api/accounts/v1/accounts"
)

func TestAccountApi(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpLogger := log.With(logger, "component", "http")

	accounts := inmem.NewAccountRepository()
	as := account.NewService(accounts)

	handler := account.MakeHandler(as, httpLogger)

	_ = accounts.Store(&account.Account{ID: "test1", Balance: decimal.NewFromFloat(1.23), Currency: "USD"})
	_ = accounts.Store(&account.Account{ID: "test2", Currency: "USD"})

	cases := []Case{
		{
			Name:   "new account:normal flow",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"id":       account.ID("asd123"),
				"balance":  decimal.NewFromFloat(99.99),
				"currency": account.CurrencyUSD,
			},
			Status:    http.StatusOK,
			Result:    CaseResponse{},
			CheckRepo: true,
		},
		{
			Name:   "new account:empty balance and currency",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"id": account.ID("asd123"),
			},
			Status:    http.StatusOK,
			Result:    CaseResponse{},
			CheckRepo: true,
		},
		{
			Name:   "new account:validation:id required",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"balance":  decimal.NewFromFloat(99.99),
				"currency": account.CurrencyUSD,
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "validation error: id: non zero value required"},
		},
		{
			Name:   "new account:validation:id alphanumeric",
			Path:   EndpointURL,
			Method: http.MethodPost,
			Payload: CaseRequestPayload{
				"balance":  decimal.NewFromFloat(99.99),
				"currency": account.CurrencyUSD,
			},
			PayloadParameter: PayloadParameter{
				Name:   "id",
				Values: []string{"abc-def", "abc_def", "фыва", "?^/"},
			},
			Status: http.StatusNotAcceptable,
			Result: CaseResponse{"error": "validation error: id: .*? does not validate as alphanum"},
		},
		{
			Name:   "delete account:normal flow",
			Path:   EndpointURL + "/asd123",
			Method: http.MethodDelete,
			Status: http.StatusOK,
			Result: CaseResponse{},
		},
		{
			Name:   "delete account:empty id",
			Path:   EndpointURL + "/",
			Method: http.MethodDelete,
			Status: http.StatusNotFound,
		},
		{
			Name:   "load account:normal flow",
			Path:   EndpointURL + "/test1",
			Method: http.MethodGet,
			Status: http.StatusOK,
			Result: CaseResponse{
				"account": CaseResponse{
					"id":       "test1",
					"balance":  1.23,
					"currency": "USD",
				},
			},
		},
		{
			Name:   "load account:empty id",
			Path:   EndpointURL + "/",
			Method: http.MethodGet,
			Status: http.StatusNotFound,
		},
		{
			Name:   "load account:wrong id",
			Path:   EndpointURL + "/qwe321",
			Method: http.MethodGet,
			Status: http.StatusNotFound,
			Result: CaseResponse{"error": "unknown account"},
		},
		{
			Name:   "load all accounts",
			Path:   EndpointURL,
			Method: http.MethodGet,
			Status: http.StatusOK,
			Result: []CaseResponse{
				{"id": "test1", "balance": 1.23, "currency": "USD"},
				{"id": "test2", "balance": 0, "currency": "USD"},
			},
		},
	}

	runTests(t, handler, cases, accounts)

	t.Run("new account:wrong json", func(t *testing.T) {
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
