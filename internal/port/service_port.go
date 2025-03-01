package port

import (
	"time"

	dbank "github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

type HelloServicePort interface {
	GenerateHello(name string) string
}

type BankServicePort interface {
	FindCurrentBalance(acct string) float64
	CreateExchangeRate(r dbank.ExchangeRate) (uuid.UUID, error)
	FindExchangeRate(fromCur string, toCur string, ts time.Time) (float64, error)
}
