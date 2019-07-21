// Package account provides handlers for work with accounts in the system.
package account

import (
	"github.com/shopspring/decimal"
)

// Currency type represents available currencies
type Currency string

const (
	CurrencyUSD Currency = "CurrencyUSD"
)

// ID type used for accounts identification.
type ID string

// Account is a wallet in the system.
type Account struct {
	ID       ID              `json:"id"`
	Balance  decimal.Decimal `json:"balance"`
	Currency Currency        `json:"currency"`
	Deleted  bool            `json:"-"`
}

// Service is the interface that provides account methods.
type Service interface {
	// New registers a new account in the system, with zero Balance.
	New(id ID, currency Currency, balance decimal.Decimal) error

	// Load returns a read model of an account.
	Load(id ID) (*Account, error)

	// LoadAll returns all accounts registered in the system.
	LoadAll() []*Account

	// Delete uses to delete account from the system. Actually mark it as deleted.
	Delete(id ID) error
}

type service struct {
	accounts Repository
}

// New registers a new account in the system, with zero Balance.
func (s *service) New(id ID, currency Currency, balance decimal.Decimal) error {
	if currency == "" {
		currency = CurrencyUSD
	}
	return s.accounts.Store(&Account{
		ID:       id,
		Balance:  balance,
		Currency: currency,
	})
}

// Load returns a read model of an account.
func (s *service) Load(id ID) (*Account, error) {
	a, err := s.accounts.Find(id)
	if err != nil {
		return nil, err
	}
	// TODO: Calculate balance as initial + per payments
	return a, nil
}

// LoadAll returns all accounts registered in the system.
func (s *service) LoadAll() []*Account {
	// TODO: Calculate balance as initial + per payments
	return s.accounts.FindAll()
}

// Delete uses to delete account from the system. Actually mark it as deleted.
func (s *service) Delete(id ID) error {
	return s.accounts.MarkDeleted(id)
}

// NewService creates an account service with necessary dependencies.
func NewService(accounts Repository) Service {
	return &service{
		accounts: accounts,
	}
}

// Repository interface for accounts storing and operations.
type Repository interface {
	// Store account in the repository
	Store(account *Account) error

	// Find account in the repository with specified id
	Find(id ID) (*Account, error)

	// FindAll returns all accounts registered in the system
	FindAll() []*Account

	// MarkDeleted is mark as deleted specified account in the system
	MarkDeleted(id ID) error
}
