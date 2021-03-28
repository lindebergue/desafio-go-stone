package router

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shopspring/decimal"

	"github.com/lindebergue/desafio-go-stone/auth"
	"github.com/lindebergue/desafio-go-stone/database"
)

// Options contains the options for creating a router.
type Options struct {
	DB        database.DB
	JWTSecret []byte
}

// New returns a new router with given opts.
func New(opts Options) http.Handler {
	h := &handler{
		db:        opts.DB,
		jwtSecret: opts.JWTSecret,
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer, middleware.RealIP, middleware.Logger)

	r.Get("/accounts", h.getAccounts)
	r.Get("/accounts/{account_id}/balance", h.getAccountBalance)
	r.Post("/accounts", h.createAccount)
	r.Post("/login", h.login)

	r.Group(func(r chi.Router) {
		r.Use(h.requireLogin)

		r.Get("/transfers", h.getTransfers)
		r.Post("/transfers", h.createTransfer)
	})

	return r
}

// handler implements the HTTP handlers for the server routes.
type handler struct {
	db        database.DB
	jwtSecret []byte
}

func (h *handler) requireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) > 7 {
			// trim the bearer prefix
			token = strings.TrimSpace(token[7:])
		}

		if token == "" {
			renderJSON(w, http.StatusForbidden, &errorResponse{Code: codeMissingBearerToken})
			return
		}

		accountID, err := auth.DecodeToken(h.jwtSecret, token)
		if err != nil {
			renderJSON(w, http.StatusForbidden, &errorResponse{Code: codeInvalidBearerToken})
			return
		}

		account, err := h.db.FindAccountByID(accountID)
		if err != nil {
			if errors.Is(err, database.ErrAccountNotFound) {
				renderJSON(w, http.StatusForbidden, &errorResponse{Code: codeInvalidBearerToken})
				return
			}
			renderServerError(w, "error finding account from bearer token: %v", err)
			return
		}

		ctx := ctxWithAccount(r.Context(), account)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) getAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.db.FindAllAccounts()
	if err != nil {
		renderServerError(w, "error finding all accounts: %v", err)
		return
	}
	renderJSON(w, http.StatusOK, accounts)
}

func (h *handler) getAccountBalance(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.ParseInt(chi.URLParam(r, "account_id"), 10, 64)
	if err != nil {
		renderJSON(w, http.StatusNotFound, &errorResponse{Code: codeAccountNotFound})
		return
	}

	account, err := h.db.FindAccountByID(accountID)
	if err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			renderJSON(w, http.StatusNotFound, &errorResponse{Code: codeAccountNotFound})
			return
		}
		renderServerError(w, "error finding account: %v", err)
		return
	}

	renderJSON(w, http.StatusOK, &balanceResponse{Balance: account.Balance})
}

func (h *handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string          `json:"name" validate:"required"`
		CPF     string          `json:"cpf" validate:"required,cpf"`
		Secret  string          `json:"secret" validate:"required,min=6"`
		Balance decimal.Decimal `json:"balance"`
	}
	if err := bindJSON(r, &body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if res := validateBody(body); res != nil {
		renderJSON(w, http.StatusUnprocessableEntity, res)
		return
	}

	hash, err := auth.HashPassword(body.Secret)
	if err != nil {
		renderServerError(w, "error hashing account secret: %v", err)
		return
	}

	account := &database.Account{
		Name:    body.Name,
		CPF:     body.CPF,
		Secret:  hash,
		Balance: body.Balance,
	}
	if err := h.db.CreateAccount(account); err != nil {
		if errors.Is(err, database.ErrAccountAlreadyExists) {
			renderJSON(w, http.StatusUnprocessableEntity, &errorResponse{Code: codeAccountAlreadyExists})
			return
		}
		renderServerError(w, "error creating account: %v", err)
		return
	}

	renderJSON(w, http.StatusCreated, account)
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CPF    string `json:"cpf"`
		Secret string `json:"secret"`
	}
	if err := bindJSON(r, &body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	account, err := h.db.FindAccountByCPF(body.CPF)
	if err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			renderJSON(w, http.StatusUnauthorized, &errorResponse{Code: codeAccountNotFound})
			return
		}
		renderServerError(w, "error finding account by cpf: %v", err)
		return
	}

	if !auth.ComparePassword(account.Secret, body.Secret) {
		renderJSON(w, http.StatusUnauthorized, &errorResponse{Code: codeAccountSecretInvalid})
		return
	}

	token, err := auth.EncodeToken(h.jwtSecret, account.ID)
	if err != nil {
		renderServerError(w, "error encoding authentication token: %v", err)
		return
	}

	renderJSON(w, http.StatusOK, &authResponse{Token: token})
}

func (h *handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	account, ok := accountFromCtx(r.Context())
	if !ok {
		renderJSON(w, http.StatusForbidden, &errorResponse{Code: codeMissingBearerToken})
		return
	}

	var body struct {
		AccountDestinationID int64           `json:"account_destination_id" validate:"required"`
		Amount               decimal.Decimal `json:"amount" validate:"required"`
	}
	if err := bindJSON(r, &body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if res := validateBody(body); res != nil {
		renderJSON(w, http.StatusUnprocessableEntity, res)
		return
	}

	transfer := &database.Transfer{
		AccountOriginID:      account.ID,
		AccountDestinationID: body.AccountDestinationID,
		Amount:               body.Amount,
	}
	if err := h.db.CreateTransfer(transfer); err != nil {
		switch {
		case errors.Is(err, database.ErrAccountNotFound):
			renderJSON(w, http.StatusUnprocessableEntity, &errorResponse{Code: codeAccountNotFound})
			return
		case errors.Is(err, database.ErrNotEnoughFunds):
			renderJSON(w, http.StatusUnprocessableEntity, &errorResponse{Code: codeAccountFundsInsuficient})
			return
		default:
			renderServerError(w, "error creating transfer: %v", err)
			return
		}
	}

	renderJSON(w, http.StatusCreated, transfer)
}

func (h *handler) getTransfers(w http.ResponseWriter, r *http.Request) {
	account, ok := accountFromCtx(r.Context())
	if !ok {
		renderJSON(w, http.StatusForbidden, &errorResponse{Code: codeMissingBearerToken})
		return
	}

	transfers, err := h.db.FindAllTransfersWithAccountID(account.ID)
	if err != nil {
		renderServerError(w, "error finding account transfers: %v", err)
		return
	}

	renderJSON(w, http.StatusOK, transfers)
}

// bindJSON binds JSON encoded data from a HTTP request into the value pointed
// by dest.
func bindJSON(r *http.Request, dest interface{}) error {
	const maxBodySize = 1024 * 1024 // 1MB
	return json.NewDecoder(io.LimitReader(r.Body, maxBodySize)).Decode(dest)
}

// renderJSON writes a HTTP JSON response with given code and body.
func renderJSON(w http.ResponseWriter, code int, body interface{}) {
	data, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// renderServerError logs an error message and renders a HTTP internal server
// error.
func renderServerError(w http.ResponseWriter, format string, args ...interface{}) {
	log.Printf(format, args...)
	w.WriteHeader(http.StatusInternalServerError)
}
