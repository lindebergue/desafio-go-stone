package database

import (
	"sync"
	"time"
)

// NewInMemDB returns a DB instance backed by local in-memory storage. Used only
// for unittests.
func NewInMemDB() DB {
	return &inmemDB{
		accounts:  map[int64]*Account{},
		transfers: map[int64]*Transfer{},
	}
}

type inmemDB struct {
	mu        sync.Mutex
	accounts  map[int64]*Account
	transfers map[int64]*Transfer
}

func (i *inmemDB) CreateAccount(account *Account) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, acc := range i.accounts {
		if acc.CPF == account.CPF {
			return ErrAccountAlreadyExists
		}
	}

	account.ID = int64(len(i.accounts) + 1)
	account.CreatedAt = time.Now()
	i.accounts[account.ID] = account
	return nil
}

func (i *inmemDB) FindAccountByID(id int64) (*Account, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, acc := range i.accounts {
		if acc.ID == id {
			return acc, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (i *inmemDB) FindAccountByCPF(cpf string) (*Account, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, acc := range i.accounts {
		if acc.CPF == cpf {
			return acc, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (i *inmemDB) FindAllAccounts() ([]*Account, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	var accounts []*Account
	for _, acc := range i.accounts {
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (i *inmemDB) CreateTransfer(transfer *Transfer) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	var (
		srcAccount *Account
		dstAccount *Account
	)
	for _, acc := range i.accounts {
		if acc.ID == transfer.AccountOriginID {
			srcAccount = acc
		}
		if acc.ID == transfer.AccountDestinationID {
			dstAccount = acc
		}
	}

	if srcAccount == nil || dstAccount == nil {
		return ErrAccountNotFound
	}
	if srcAccount.Balance.LessThan(transfer.Amount) {
		return ErrNotEnoughFunds
	}

	transfer.ID = int64(len(i.transfers) + 1)
	srcAccount.Balance = srcAccount.Balance.Sub(transfer.Amount)
	dstAccount.Balance = dstAccount.Balance.Add(transfer.Amount)
	i.accounts[srcAccount.ID] = srcAccount
	i.accounts[dstAccount.ID] = dstAccount
	i.transfers[transfer.ID] = transfer

	return nil
}

func (i *inmemDB) FindAllTransfersWithAccountID(accountID int64) ([]*Transfer, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	var transfers []*Transfer
	for _, t := range i.transfers {
		if t.AccountOriginID == accountID || t.AccountDestinationID == accountID {
			transfers = append(transfers, t)
		}
	}

	return transfers, nil
}
