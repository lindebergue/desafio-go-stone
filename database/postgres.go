package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"

	_ "github.com/lindebergue/desafio-go-stone/database/migrations" // load migrations
)

// NewPostgresDB returns a DB instance backed by a Postgres database specified
// by url.
// URLs use the format postgres://user:password@host:port/database?option=value.
// Refer to the go-pg package documentation for the complete URL syntax.
//
// We use an exponential backoff for testing the connection in order to avoid
// scenarios where the server is ready to accept connections but the
// database still not up (like docker-compose environments).
func NewPostgresDB(url string) (DB, error) {
	opts, err := pg.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	db := pg.Connect(opts)

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxElapsedTime = 15 * time.Second

	retryFn := func() error {
		err := db.Ping(context.Background())
		if _, ok := err.(pg.Error); ok {
			return backoff.Permanent(err)
		}
		return err
	}
	if err := backoff.Retry(retryFn, b); err != nil {
		return nil, err
	}

	if _, _, err := migrations.Run(db, "init"); err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	if _, _, err := migrations.Run(db, "up"); err != nil {
		return nil, fmt.Errorf("error migrating database: %w", err)
	}

	return &postgresDB{db: db}, nil
}

type postgresDB struct {
	db *pg.DB
}

func (p *postgresDB) CreateAccount(account *Account) error {
	_, err := p.db.Model(account).
		Column("name", "cpf", "secret", "balance").
		Returning("*").
		Insert()

	if pgErr, ok := err.(pg.Error); ok && pgErr.Field('C') == "23505" {
		return ErrAccountAlreadyExists
	}
	return wrapPostgresError(err)
}

func (p *postgresDB) FindAccountByID(id int64) (*Account, error) {
	account := &Account{}
	err := p.db.Model(account).
		Where("account.id = ?", id).
		Select()

	return account, wrapPostgresError(err)
}

func (p *postgresDB) FindAccountByCPF(cpf string) (*Account, error) {
	account := &Account{}
	err := p.db.Model(account).
		Where("account.cpf = ?", cpf).
		Select()

	return account, wrapPostgresError(err)
}

func (p *postgresDB) FindAllAccounts() ([]*Account, error) {
	var accounts []*Account
	err := p.db.Model(&accounts).
		Order("account.created_at ASC").
		Select()

	return accounts, wrapPostgresError(err)
}

func (p *postgresDB) CreateTransfer(transfer *Transfer) error {
	err := p.db.RunInTransaction(context.Background(), func(t *pg.Tx) error {
		srcAccount := &Account{}
		if err := t.Model(srcAccount).Where("account.id = ?", transfer.AccountOriginID).Select(); err != nil {
			return err
		}

		dstAccount := &Account{}
		if err := t.Model(dstAccount).Where("account.id = ?", transfer.AccountDestinationID).Select(); err != nil {
			return err
		}

		if srcAccount.Balance.LessThan(transfer.Amount) {
			return ErrNotEnoughFunds
		}

		srcAccount.Balance = srcAccount.Balance.Sub(transfer.Amount)
		dstAccount.Balance = dstAccount.Balance.Add(transfer.Amount)

		if _, err := t.Model(srcAccount).WherePK().Update(); err != nil {
			return err
		}
		if _, err := t.Model(dstAccount).WherePK().Update(); err != nil {
			return err
		}

		_, err := t.Model(transfer).
			Column("account_origin_id", "account_destination_id", "amount").
			Returning("*").
			Insert()

		return err
	})
	return wrapPostgresError(err)
}

func (p *postgresDB) FindAllTransfersWithAccountID(accountID int64) ([]*Transfer, error) {
	var transfers []*Transfer
	err := p.db.Model(&transfers).
		Where("transfer.account_origin_id = ?", accountID).
		WhereOr("transfer.account_destination_id = ?", accountID).
		Order("transfer.created_at DESC").
		Select()

	return transfers, wrapPostgresError(err)
}

func wrapPostgresError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pg.ErrNoRows):
		return ErrAccountNotFound
	default:
		if _, ok := err.(pg.Error); ok {
			return fmt.Errorf("internal database error: %w", err)
		}
		return err
	}
}
