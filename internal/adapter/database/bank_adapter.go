package database

import (
	"log"
	"time"

	dbank "github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

func (a *DatabaseAdapter) GetBankAccountByAccountNumber(acct string) (BankAccountOrm, error) {
	var BankAccountOrm BankAccountOrm

	if err := a.db.First(&BankAccountOrm, "account_number = ?", acct).Error; err != nil {
		log.Printf("Can't find bank account number %v : %v\n", acct, err)
		return BankAccountOrm, err
	}

	return BankAccountOrm, nil
}

func (a *DatabaseAdapter) CreateExchangeRate(r BankExchangeRateOrm) (uuid.UUID, error) {
	if err := a.db.Create(r).Error; err != nil {
		return uuid.Nil, err
	}

	return r.ExchangeRateUuid, nil
}

func (a *DatabaseAdapter) GetExchangeRateAtTimestamp(fromCur string, toCur string,
	ts time.Time) (BankExchangeRateOrm, error) {
	var exchangeRateOrm BankExchangeRateOrm

	err := a.db.First(&exchangeRateOrm, "from_currency = ? "+" AND to_currency = ? "+
		" AND (? BETWEEN valid_from_timestamp and valid_to_timestamp)", fromCur, toCur, ts).Error

	return exchangeRateOrm, err
}

func (a *DatabaseAdapter) CreateTransaction(acct BankAccountOrm, t BankTransactionOrm) (uuid.UUID, error) {
	tx := a.db.Begin()

	if err := tx.Create(t).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	// recalculate current balance
	newAmount := t.Amount

	if t.TransactionType == dbank.TransactionTypeOut {
		newAmount = -1 * t.Amount
	}

	newAccountBalance := acct.CurrentBalance + newAmount

	if err := tx.Model(&acct).Updates(
		map[string]interface{}{
			"current_balance": newAccountBalance,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	tx.Commit()

	return t.TransactionUuid, nil
}
