// Package database implements the data access layer.
package database

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	// ErrAccountNotFound indicates that an account cannot be found.
	ErrAccountNotFound = errors.New("database: account not found")

	// ErrAccountAlreadyExists indicates that an account with same CPF already
	// exists.
	ErrAccountAlreadyExists = errors.New("database: account already exists")

	// ErrNotEnoughFunds indicates that a transfer cannot be completed due to
	// insufficient funds.
	ErrNotEnoughFunds = errors.New("database: not enough funds")
)

// Account represents a bank account and its balance.
type Account struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	CPF       string          `json:"cpf"`
	Secret    string          `json:"-"`
	Balance   decimal.Decimal `json:"balance" pg:",use_zero"`
	CreatedAt time.Time       `json:"created_at"`
}

// Transfer represents a balance transfer between accounts.
type Transfer struct {
	ID                   int64           `json:"id"`
	AccountOriginID      int64           `json:"account_origin_id"`
	AccountDestinationID int64           `json:"account_destination_id"`
	Amount               decimal.Decimal `json:"amount" pg:",use_zero"`
	CreatedAt            time.Time       `json:"created_at"`
}

// DB provides methods for managing application data.
type DB interface {
	// CreateAccount adds an account into the database. Returns
	// ErrAccountAlreadyExists if already exists an account with same CPF.
	CreateAccount(account *Account) error

	// FindAccountByID finds an account by its ID. Returns ErrAccountNotFound
	// if the account cannot be found.
	FindAccountByID(id int64) (*Account, error)

	// FindAccountByCPF finds an account by its CPF. Returns ErrAccountNotFound
	// if the account cannot be found.
	FindAccountByCPF(cpf string) (*Account, error)

	// FindAllAccounts finds all accounts from the database.
	FindAllAccounts() ([]*Account, error)

	// CreateTransfer creates a transfer between two accounts, adjusting their
	// balances accordingly. If the origin account does not have enough funds,
	// returns ErrNotEnoughFunds. If any of the accounts of the operation does
	// not exist, returns ErrAccountNotFound.
	CreateTransfer(transfer *Transfer) error

	// FindAllTransfersWithAccountId finds all transfers with accountID as origin or
	// destination.
	FindAllTransfersWithAccountID(accountID int64) ([]*Transfer, error)
}
