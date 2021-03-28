package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lindebergue/desafio-go-stone/database"
)

func TestRouter(t *testing.T) {
	router := New(Options{
		DB: database.NewInMemDB(database.WithNowFunc(func() time.Time {
			return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		})),
		JWTSecret: []byte("secret"),
	})
	tests := []struct {
		testcase         string
		method           string
		path             string
		headers          map[string]string
		body             string
		expectedStatus   int
		expectedResponse string
	}{
		{
			testcase: "create first account",
			method:   "POST",
			path:     "/accounts",
			body: `
				{
					"name": "first account",
					"cpf": "111.111.111-11",
					"secret": "firstsecret",
					"balance": "100"
				}
			`,
			expectedStatus: http.StatusCreated,
		},
		{
			testcase: "create second account",
			method:   "POST",
			path:     "/accounts",
			body: `
				{
					"name": "second account",
					"cpf": "222.222.222-22",
					"secret": "secondsecret",
					"balance": "50"
				}
			`,
			expectedStatus: http.StatusCreated,
		},
		{
			testcase: "create account with same cpf",
			method:   "POST",
			path:     "/accounts",
			body: `
				{
					"name": "other account with same cpf",
					"cpf": "111.111.111-11",
					"secret": "othersecret",
					"balance": "0"
				}
			`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: `
				{
					"code": "ACCOUNT_ALREADY_EXISTS"
				}
			`,
		},
		{
			testcase:       "list accounts",
			method:         "GET",
			path:           "/accounts",
			expectedStatus: http.StatusOK,
			expectedResponse: `
				[
					{
						"id": 1,
						"name": "first account",
						"cpf": "111.111.111-11",
						"balance": "100",
						"created_at": "2021-01-01T00:00:00Z"
					},
					{
						"id": 2,
						"name": "second account",
						"cpf": "222.222.222-22",
						"balance": "50",
						"created_at": "2021-01-01T00:00:00Z"
					}
				]
			`,
		},
		{
			testcase:       "get account balance",
			method:         "GET",
			path:           "/accounts/1/balance",
			expectedStatus: http.StatusOK,
			expectedResponse: `
				{
					"balance": "100"
				}
			`,
		},
		{
			testcase:       "get balance from account that does not exist",
			method:         "GET",
			path:           "/accounts/1000/balance",
			expectedStatus: http.StatusNotFound,
			expectedResponse: `
				{
					"code": "ACCOUNT_NOT_FOUND"
				}
			`,
		},
		{
			testcase: "login",
			method:   "POST",
			path:     "/login",
			body: `
				{
					"cpf": "111.111.111-11",
					"secret": "firstsecret"
				}
			`,
			expectedStatus: http.StatusOK,
		},
		{
			testcase: "login with wrong secret",
			method:   "POST",
			path:     "/login",
			body: `
				{
					"cpf": "111.111.111-11",
					"secret": "wrongsecret"
				}
			`,
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: `
				{
					"code": "ACCOUNT_SECRET_INVALID"
				}
			`,
		},
		{
			testcase: "login with wrong cpf",
			method:   "POST",
			path:     "/login",
			body: `
				{
					"cpf": "999.999.999-99",
					"secret": "somesecret"
				}
			`,
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: `
				{
					"code": "ACCOUNT_NOT_FOUND"
				}
			`,
		},
		{
			testcase: "transfer",
			method:   "POST",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsImh0dHBzOi8vZGVzYWZpby1nby1zdG9uZS5sb2NhbC9hY2NvdW50X2lkIjoxfQ.mTvJjwO1SCYBVvailOd1kwM78fyuiOHbzptDouDNf1g",
			},
			body: `
				{
					"account_destination_id": 2,
					"amount": 0.1
				}
			`,
			expectedStatus: http.StatusCreated,
		},
		{
			testcase: "transfer without jwt",
			method:   "POST",
			path:     "/transfers",
			body: `
				{
					"account_destination_id": 2,
					"amount": 0.1
				}
			`,
			expectedStatus: http.StatusForbidden,
			expectedResponse: `
				{
					"code": "MISSING_BEARER_TOKEN"
				}
			`,
		},
		{
			testcase: "transfer with malformed jwt",
			method:   "POST",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "bearer notavalidjwt",
			},
			body: `
				{
					"account_destination_id": 2,
					"amount": 0.1
				}
			`,
			expectedStatus: http.StatusForbidden,
			expectedResponse: `
				{
					"code": "INVALID_BEARER_TOKEN"
				}
			`,
		},
		{
			testcase: "transfer with a jwt for an account that does not exist",
			method:   "POST",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsImh0dHBzOi8vZGVzYWZpby1nby1zdG9uZS5sb2NhbC9hY2NvdW50X2lkIjoxMDB9.8vOPLzkyDkbKQjFf8nzWbDNW3RUerO9ulgMGDuR8VbU",
			},
			body: `
				{
					"account_destination_id": 2,
					"amount": 0.1
				}
			`,
			expectedStatus: http.StatusForbidden,
			expectedResponse: `
				{
					"code": "INVALID_BEARER_TOKEN"
				}
			`,
		},
		{
			testcase: "transfer without enough funds",
			method:   "POST",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsImh0dHBzOi8vZGVzYWZpby1nby1zdG9uZS5sb2NhbC9hY2NvdW50X2lkIjoxfQ.mTvJjwO1SCYBVvailOd1kwM78fyuiOHbzptDouDNf1g",
			},
			body: `
				{
					"account_destination_id": 2,
					"amount": 1000
				}
			`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: `
				{
					"code": "ACCOUNT_FUNDS_INSUFICIENT"
				}
			`,
		},
		{
			testcase: "transfer to invalid account id",
			method:   "POST",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsImh0dHBzOi8vZGVzYWZpby1nby1zdG9uZS5sb2NhbC9hY2NvdW50X2lkIjoxfQ.mTvJjwO1SCYBVvailOd1kwM78fyuiOHbzptDouDNf1g",
			},
			body: `
				{
					"account_destination_id": 1000,
					"amount": 10
				}
			`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: `
				{
					"code": "ACCOUNT_NOT_FOUND"
				}
			`,
		},
		{
			testcase: "get transfers",
			method:   "GET",
			path:     "/transfers",
			headers: map[string]string{
				"authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsImh0dHBzOi8vZGVzYWZpby1nby1zdG9uZS5sb2NhbC9hY2NvdW50X2lkIjoxfQ.mTvJjwO1SCYBVvailOd1kwM78fyuiOHbzptDouDNf1g",
			},
			expectedStatus: http.StatusOK,
			expectedResponse: `
				[
					{
						"id": 1,
						"account_origin_id": 1,
						"account_destination_id": 2,
						"amount": "0.1",
						"created_at": "2021-01-01T00:00:00Z"
					}
				]
			`,
		},
	}

	for _, test := range tests {
		t.Run(test.testcase, func(t *testing.T) {

			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))

			if test.body != "" {
				r.Header.Set("content-type", "application/json")
			}
			for key, value := range test.headers {
				r.Header.Set(key, value)
			}

			router.ServeHTTP(w, r)

			assert.Equal(t, test.expectedStatus, w.Code)
			if test.expectedResponse != "" {
				assert.JSONEq(t, test.expectedResponse, w.Body.String())
			}
		})
	}
}
