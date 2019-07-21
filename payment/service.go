package payment

import (
	"github.com/google/uuid"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/errs"
	"github.com/shopspring/decimal"
)

// Direction of payment regarding account.
type Direction string

const (
	Incoming Direction = "incoming"
	Outgoing Direction = "outgoing"
)

// Payment holding a money transfer between two accounts in the system.
type Payment struct {
	ID          uuid.UUID       `json:"-"`
	Account     account.ID      `json:"account"`
	Amount      decimal.Decimal `json:"amount"`
	ToAccount   account.ID      `json:"to_account,omitempty"`
	FromAccount account.ID      `json:"from_account,omitempty"`
	Direction   Direction       `json:"direction"`
	Deleted     bool            `json:"-"`
}

// Service is the interface that provides payment methods.
type Service interface {
	// New registers a new payment in the system.
	New(fromAccountID account.ID, amount decimal.Decimal, toAccountID account.ID) error

	// Load returns payments list for an account.
	Load(accountID account.ID) []*Payment

	// LoadAll returns all payments, registered in the system.
	LoadAll() []*Payment
}

type service struct {
	accounts account.Repository
	payments Repository
}

// New registers a new payment in the system.
func (s *service) New(fromAccountID account.ID, amount decimal.Decimal, toAccountID account.ID) error {
	if fromAccountID == toAccountID {
		return errs.ErrAccountsAreEqual
	}
	from, err := s.accounts.Find(fromAccountID)
	if err != nil {
		return errs.ErrUnknownSourceAccount
	}
	if from.Balance.LessThan(amount) { // TODO: Balance need to be calculating property
		return errs.ErrInsufficientMoney
	}
	_, err = s.accounts.Find(toAccountID)
	if err != nil {
		return errs.ErrUnknownTargetAccount
	}

	outgoingPayment := Payment{
		ID:        uuid.New(),
		Account:   fromAccountID,
		Amount:    amount,
		ToAccount: toAccountID,
		Direction: Outgoing,
	}
	incomingPayment := Payment{
		ID:          uuid.New(),
		Account:     toAccountID,
		Amount:      amount,
		FromAccount: fromAccountID,
		Direction:   Incoming,
	}
	err = s.payments.Store(&outgoingPayment, &incomingPayment)
	if err != nil {
		return errs.ErrStorePayments
	}
	return nil
}

// Load returns payments list for an account.
func (s *service) Load(accountID account.ID) []*Payment {
	return s.payments.Find(accountID)
}

// LoadAll returns all payments, registered in the system.
func (s *service) LoadAll() []*Payment {
	return s.payments.FindAll()
}

// NewService creates a payment service with necessary dependencies.
func NewService(payments Repository, accounts account.Repository) Service {
	return &service{
		payments: payments,
		accounts: accounts,
	}
}

// Repository interface for payment storing and operations.
type Repository interface {
	// Store payment in the repository.
	Store(payment ...*Payment) error

	// Find payments list for an account.
	Find(id account.ID) []*Payment

	// FindAll returns all payments, registered in the system.
	FindAll() []*Payment

	// MarkDeleted is mark as deleted specified payment in the system
	MarkDeleted(id uuid.UUID) error
}
