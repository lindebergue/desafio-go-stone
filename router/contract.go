package router

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"github.com/lindebergue/desafio-go-stone/database"
)

// contextKey is the type of context keys.
type contextKey string

// The context keys for data that transits between HTTP handlers.
var (
	accountContextKey contextKey = "github.com/lindebergue/desafio-go-stone/account"
)

// errorResponse represents a server error response.
type errorResponse struct {
	Code    errorCode `json:"code"`
	Details string    `json:"details,omitempty"`
}

// accountFromCtx returns the account stored into a context value.
func accountFromCtx(ctx context.Context) (*database.Account, bool) {
	account, ok := ctx.Value(accountContextKey).(*database.Account)
	return account, ok
}

// ctxWithAccount returns a copy of ctx with account stored in.
func ctxWithAccount(ctx context.Context, account *database.Account) context.Context {
	return context.WithValue(ctx, accountContextKey, account)
}

// errorCode represents the code of an error response.
type errorCode string

// The error codes.
const (
	codeMissingBearerToken      errorCode = "MISSING_BEARER_TOKEN"
	codeInvalidBearerToken      errorCode = "INVALID_BEARER_TOKEN"
	codeValidationError         errorCode = "VALIDATION_ERROR"
	codeAccountAlreadyExists    errorCode = "ACCOUNT_ALREADY_EXISTS"
	codeAccountNotFound         errorCode = "ACCOUNT_NOT_FOUND"
	codeAccountSecretInvalid    errorCode = "ACCOUNT_SECRET_INVALID"
	codeAccountFundsInsuficient errorCode = "ACCOUNT_FUNDS_INSUFICIENT"
)

// balanceResponse represents the response of an account balance.
type balanceResponse struct {
	Balance decimal.Decimal `json:"balance"`
}

// authResponse represents the response of a successful login.
type authResponse struct {
	Token string `json:"token"`
}

var validate = validator.New()

func init() {
	validate.RegisterValidation("cpf", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`\d{3}\.\d{3}\.\d{3}\-\d{2}`).MatchString(fl.Field().String())
	})
}

// validateBody validates the body of a HTTP request.
func validateBody(data interface{}) *errorResponse {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	var msgs []string
	for _, ve := range err.(validator.ValidationErrors) {
		switch ve.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", strings.ToLower(ve.Field())))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must have at least %s characters", strings.ToLower(ve.Field()), ve.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is not valid", strings.ToLower(ve.Field())))
		}
	}

	return &errorResponse{
		Code:    codeValidationError,
		Details: strings.Join(msgs, ";"),
	}
}
