package database

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func runDBTests(t *testing.T, db DB) {
	acc1 := &Account{
		Name:    "first account",
		CPF:     "111.111.111-11",
		Secret:  "verysecretpassword",
		Balance: decimal.NewFromFloat(0.1),
	}
	acc2 := &Account{
		Name:    "second account",
		CPF:     "222.222.222-22",
		Secret:  "othersecretpassword",
		Balance: decimal.NewFromFloat(0.2),
	}

	t.Run("create", func(t *testing.T) {
		require.NoError(t, db.CreateAccount(acc1))
		require.NoError(t, db.CreateAccount(acc2))
		require.NotEmpty(t, acc1.ID)
		require.NotEmpty(t, acc2.ID)
	})

	t.Run("create with same cpf", func(t *testing.T) {
		acc3 := &Account{
			Name:    "duplicated account",
			CPF:     acc1.CPF,
			Secret:  "secret",
			Balance: decimal.Zero,
		}
		require.Equal(t, ErrAccountAlreadyExists, db.CreateAccount(acc3))
	})

	t.Run("find", func(t *testing.T) {
		foundByID, err := db.FindAccountByID(acc1.ID)
		require.NoError(t, err)
		require.Equal(t, acc1, foundByID)

		foundByCPF, err := db.FindAccountByCPF(acc1.CPF)
		require.NoError(t, err)
		require.Equal(t, acc1, foundByCPF)

		foundAll, err := db.FindAllAccounts()
		require.Len(t, foundAll, 2)
	})

	t.Run("find accounts that does not exist", func(t *testing.T) {
		_, err := db.FindAccountByID(42)
		require.Equal(t, ErrAccountNotFound, err)

		_, err = db.FindAccountByCPF("000.000.000-00")
		require.Equal(t, ErrAccountNotFound, err)
	})

	t.Run("transfer", func(t *testing.T) {
		transf := &Transfer{
			AccountOriginID:      acc1.ID,
			AccountDestinationID: acc2.ID,
			Amount:               acc1.Balance,
		}
		require.NoError(t, db.CreateTransfer(transf))

		src, err := db.FindAccountByID(acc1.ID)
		require.NoError(t, err)
		require.True(t, src.Balance.IsZero())

		dst, err := db.FindAccountByID(acc2.ID)
		require.NoError(t, err)
		require.True(t, dst.Balance.Equal(decimal.NewFromFloat(0.3)))

		transfers, err := db.FindAllTransfersWithAccountID(acc1.ID)
		require.NotEmpty(t, transfers)
	})

	t.Run("transfer without enough funds", func(t *testing.T) {
		transf := &Transfer{
			AccountOriginID:      acc1.ID,
			AccountDestinationID: acc2.ID,
			Amount:               decimal.NewFromFloat(1_000_000),
		}
		require.Equal(t, ErrNotEnoughFunds, db.CreateTransfer(transf))
	})

	t.Run("transfer from account that does not exist", func(t *testing.T) {
		transf := &Transfer{
			AccountOriginID:      1000,
			AccountDestinationID: acc2.ID,
			Amount:               decimal.NewFromFloat(1),
		}
		require.Equal(t, ErrAccountNotFound, db.CreateTransfer(transf))
	})

	t.Run("transfer to an account that does not exist", func(t *testing.T) {
		transf := &Transfer{
			AccountOriginID:      acc1.ID,
			AccountDestinationID: 1001,
			Amount:               decimal.NewFromFloat(1),
		}
		require.Equal(t, ErrAccountNotFound, db.CreateTransfer(transf))
	})

	t.Run("transfer between accounts that does not exist", func(t *testing.T) {
		transf := &Transfer{
			AccountOriginID:      1000,
			AccountDestinationID: 1001,
			Amount:               decimal.NewFromFloat(1),
		}
		require.Equal(t, ErrAccountNotFound, db.CreateTransfer(transf))
	})
}
