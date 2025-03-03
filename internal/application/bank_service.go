package application

import (
	"fmt"
	"log"
	"time"

	db "github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/adapter/database"
	"github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/application/domain/bank"
	dbank "github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/application/domain/bank"
	"github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/port"
	"github.com/google/uuid"
)

type BankService struct {
	db port.BankDatabasePort
}

func NewBankService(dbPort port.BankDatabasePort) *BankService {
	return &BankService{db: dbPort}
}

func (s *BankService) FindCurrentBalance(acct string) float64 {
	bankAccount, err := s.db.GetBankAccountByAccountNumber(acct)

	if err != nil {
		log.Println("Error on FindCurrentBalance", err)
	}

	return bankAccount.CurrentBalance
}

func (s *BankService) CreateExchangeRate(r dbank.ExchangeRate) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	exchangeRateOrm := db.BankExchangeRateOrm{
		ExchangeRateUuid:   newUuid,
		FromCurrency:       r.FromCurrency,
		ToCurrency:         r.ToCurrency,
		Rate:               r.Rate,
		ValidFromTimestamp: r.ValidFromTimestamp,
		ValidToTimestamp:   r.ValidToTimestamp,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	return s.db.CreateExchangeRate(exchangeRateOrm)
}

func (s *BankService) FindExchangeRate(fromCur string, toCur string, ts time.Time) (float64, error) {
	exchangeRate, err := s.db.GetExchangeRateAtTimestamp(fromCur, toCur, ts)

	if err != nil {
		return 0, err
	}

	return float64(exchangeRate.Rate), nil
}

func (s *BankService) CreateTransaction(acct string, t dbank.Transaction) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	bankAccountOrm, err := s.db.GetBankAccountByAccountNumber(acct)

	if err != nil {
		log.Printf("Can't create transaction for %v : %v\n", acct, err)
		return uuid.Nil, fmt.Errorf("can't find account number %v : %v", acct, err.Error())
	}

	if t.TransactionType == bank.TransactionTypeOut && bankAccountOrm.CurrentBalance < t.Amount {
		return bankAccountOrm.AccountUuid, fmt.Errorf(
			"insufficient account balance %v for [out] transaction amount %v",
			bankAccountOrm.CurrentBalance, t.Amount,
		)
	}

	transactionOrm := db.BankTransactionOrm{
		TransactionUuid:      newUuid,
		AccountUuid:          bankAccountOrm.AccountUuid,
		TransactionTimestamp: now,
		Amount:               t.Amount,
		TransactionType:      t.TransactionType,
		Notes:                t.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	savedUuid, err := s.db.CreateTransaction(bankAccountOrm, transactionOrm)

	return savedUuid, err
}

func (s *BankService) CalculateTransactionSummary(tcur *dbank.TransactionSummary,
	trans dbank.Transaction) error {
	switch trans.TransactionType {
	case dbank.TransactionTypeIn:
		tcur.SumIn += trans.Amount
	case dbank.TransactionTypeOut:
		tcur.SumOut += trans.Amount
	default:
		return fmt.Errorf("unknown transaction type %v", trans.TransactionType)
	}

	tcur.SumTotal = tcur.SumIn - tcur.SumOut

	return nil
}
