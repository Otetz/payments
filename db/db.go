// Package db provide PostgreSQL storage for application data repositories.
package db

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/google/uuid"
	"github.com/otetz/payments/account"
	"github.com/otetz/payments/errs"
	"github.com/otetz/payments/payment"
)

// CreateSchema creating schema if its not exist. Without any migrations mechanic, just schema only.
func CreateSchema(conn *pg.DB) error {
	for _, model := range []interface{}{(*account.Account)(nil), (*payment.Payment)(nil)} {
		err := conn.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type accountRepository struct {
	conn *pg.DB
}

// Store account in the repository
func (r *accountRepository) Store(account *account.Account) error {
	if err := r.conn.Insert(account); err != nil {
		return err
	}
	return nil
}

// Find account in the repository with specified id
func (r *accountRepository) Find(id account.ID) (*account.Account, error) {
	a := &account.Account{ID: id}
	err := r.conn.Select(a)
	if err != nil {
		return nil, err
	}
	if a.Deleted {
		return nil, errs.ErrUnknownAccount
	}
	return a, nil
}

// FindAll returns all accounts registered in the system
func (r *accountRepository) FindAll() []*account.Account {
	var accounts []*account.Account
	err := r.conn.Model(&accounts).Where("deleted = ?", false).Select()
	if err != nil {
		return nil
	}
	return accounts
}

// MarkDeleted is mark as deleted specified account in the system
func (r *accountRepository) MarkDeleted(id account.ID) error {
	a := &account.Account{ID: id}
	err := r.conn.Select(a)
	if err != nil {
		return err
	}
	if a.ID == "" || a.Deleted {
		return errs.ErrUnknownAccount
	}
	a.Deleted = true
	err = r.conn.Update(a)
	if err != nil {
		return err
	}
	return nil
}

// NewAccountRepository returns a new instance of a PostgreSQL account repository.
func NewAccountRepository(conn *pg.DB) account.Repository {
	return &accountRepository{
		conn: conn,
	}
}

type paymentRepository struct {
	conn     *pg.DB
	accounts account.Repository
}

// Store payments in the repository.
func (r *paymentRepository) Store(payments ...*payment.Payment) error {
	err := r.conn.RunInTransaction(func(tx *pg.Tx) error {
		for _, val := range payments {
			if err := r.conn.Insert(val); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Find payments list for an account.
func (r *paymentRepository) Find(id account.ID) []*payment.Payment {
	var pp []*payment.Payment
	err := r.conn.Model(&pp).Where("deleted = ?", false).Where("account = ?", id).Select()
	if err != nil {
		return nil
	}
	return pp
}

// FindAll returns all payments, registered in the system.
func (r *paymentRepository) FindAll() []*payment.Payment {
	var pp []*payment.Payment
	err := r.conn.Model(&pp).Where("deleted = ?", false).Select()
	if err != nil {
		return nil
	}
	return pp
}

// MarkDeleted is mark as deleted specified payment in the system
func (r *paymentRepository) MarkDeleted(id uuid.UUID) error {
	p := &payment.Payment{ID: id}
	err := r.conn.Select(p)
	if err != nil {
		return err
	}
	p.Deleted = true
	err = r.conn.Update(p)
	if err != nil {
		return err
	}
	return nil
}

// NewPaymentRepository returns a new instance of a PostgreSQL payment repository.
func NewPaymentRepository(conn *pg.DB, accounts account.Repository) payment.Repository {
	return &paymentRepository{
		conn:     conn,
		accounts: accounts,
	}
}
