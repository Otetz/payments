// Package inmem provide in-memory storage for application data repositories.
package inmem

import (
	"sync"

	"github.com/google/uuid"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/errs"
	"github.com/otetz/payments/payment"
)

type accountRepository struct {
	mtx      sync.RWMutex
	accounts map[account.ID]*account.Account
}

// Store account in the repository
func (r *accountRepository) Store(c *account.Account) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.accounts[c.ID] = c
	return nil
}

// Find account in the repository with specified id
func (r *accountRepository) Find(id account.ID) (*account.Account, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	if val, ok := r.accounts[id]; ok {
		if !val.Deleted {
			return val, nil
		}
	}
	return nil, errs.ErrUnknownAccount
}

// FindAll returns all accounts registered in the system
func (r *accountRepository) FindAll() []*account.Account {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	c := make([]*account.Account, 0, len(r.accounts))
	for _, val := range r.accounts {
		if !val.Deleted {
			c = append(c, val)
		}
	}
	return c
}

// MarkDeleted is mark as deleted specified account in the system
func (r *accountRepository) MarkDeleted(id account.ID) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.accounts[id]; ok {
		//delete(r.accounts, id)
		r.accounts[id].Deleted = true
		return nil
	}
	return errs.ErrUnknownAccount
}

// NewAccountRepository returns a new instance of an in-memory account repository.
func NewAccountRepository() account.Repository {
	return &accountRepository{
		accounts: make(map[account.ID]*account.Account),
	}
}

type paymentRepository struct {
	mtx      sync.RWMutex
	payments map[uuid.UUID]*payment.Payment
}

// Store payment in the repository.
func (r *paymentRepository) Store(p *payment.Payment) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.payments[p.ID] = p
	return nil
}

// Find payments list for an account.
func (r *paymentRepository) Find(id account.ID) []*payment.Payment {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	result := make([]*payment.Payment, 0, len(r.payments))
	for _, val := range r.payments {
		if val.Account == id && !val.Deleted {
			result = append(result, val)
		}
	}
	return result
}

// FindAll returns all payments, registered in the system.
func (r *paymentRepository) FindAll() []*payment.Payment {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	result := make([]*payment.Payment, 0, len(r.payments))
	for _, val := range r.payments {
		if !val.Deleted {
			result = append(result, val)
		}
	}
	return result
}

// NewPaymentRepository returns a new instance of an in-memory payment repository.
func NewPaymentRepository() payment.Repository {
	return &paymentRepository{
		payments: make(map[uuid.UUID]*payment.Payment),
	}
}
