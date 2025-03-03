package grpc

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/AlfianVitoAnggoro/my-grpc-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/grpc"

	dbank "github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/application/domain/bank"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank.CurrentBalanceRequest) (*bank.CurrentBalanceResponse, error) {
	now := time.Now()
	bal := a.bankService.FindCurrentBalance(req.AccountNumber)

	return &bank.CurrentBalanceResponse{
		Amount: bal,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}

func (a *GrpcAdapter) FetchExchangeRates(req *bank.ExchangeRateRequest, stream grpc.ServerStreamingServer[bank.ExchangeRateResponse]) error {
	context := stream.Context()

	for {
		select {
		case <-context.Done():
			log.Println("Client cancelled stream")
			return nil

		default:
			now := time.Now().Truncate(time.Second)
			rate, err := a.bankService.FindExchangeRate(req.FromCurrency, req.ToCurrency, now)

			if err != nil {
				return err
			}

			stream.Send(
				&bank.ExchangeRateResponse{
					FromCurrency: req.FromCurrency,
					ToCurrency:   req.ToCurrency,
					Rate:         rate,
					Timestamp:    now.Format(time.RFC3339),
				},
			)

			log.Printf("Exchange rate sent to client, %v to %v : %v\n", req.FromCurrency, req.ToCurrency, rate)

			time.Sleep(3 * time.Second)
		}
	}
}

func toTime(dt *datetime.DateTime) (time.Time, error) {
	if dt == nil {
		now := time.Now()

		dt = &datetime.DateTime{
			Year:    int32(now.Year()),
			Month:   int32(now.Month()),
			Day:     int32(now.Day()),
			Hours:   int32(now.Hour()),
			Minutes: int32(now.Minute()),
			Seconds: int32(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		}
	}

	res := time.Date(int(dt.Year), time.Month(dt.Month), int(dt.Day),
		int(dt.Hours), int(dt.Minutes), int(dt.Seconds), int(dt.Nanos), time.UTC)

	return res, nil
}

func (a *GrpcAdapter) SummarizeTransactions(stream bank.BankService_SummarizeTransactionsServer) error {
	tsum := dbank.TransactionSummary{
		SummaryOnDate: time.Now(),
		SumIn:         0,
		SumOut:        0,
		SumTotal:      0,
	}
	acct := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			res := bank.TransactionSummary{
				AccountNumber: acct,
				SumAmountIn:   tsum.SumIn,
				SumAmountOut:  tsum.SumOut,
				SumTotal:      tsum.SumTotal,
				TransactionDate: &date.Date{
					Year:  int32(tsum.SummaryOnDate.Year()),
					Month: int32(tsum.SummaryOnDate.Month()),
					Day:   int32(tsum.SummaryOnDate.Day()),
				},
			}

			return stream.SendAndClose(&res)
		}

		if err != nil {
			log.Fatalln("Error while reading from client :", err)
		}

		acct = req.AccountNumber
		ts, err := toTime(req.Timestamp)

		if err != nil {
			log.Fatalf("Error while parsing timestamp %v : %v", req.Timestamp, err)
		}

		ttype := dbank.TransactionTypeUnknown

		if req.Type == bank.TransactionType_TRANSACTION_TYPE_IN {
			ttype = dbank.TransactionTypeIn
		} else if req.Type == bank.TransactionType_TRANSACTION_TYPE_OUT {
			ttype = dbank.TransactionTypeOut
		}

		tcur := dbank.Transaction{
			Amount:          req.Amount,
			Timestamp:       ts,
			TransactionType: ttype,
		}

		_, err = a.bankService.CreateTransaction(req.AccountNumber, tcur)

		// if err != nil && accountUuid == uuid.Nil {
		// 	s := status.New(codes.InvalidArgument, err.Error())
		// 	s, _ = s.WithDetails(&errdetails.BadRequest{
		// 		FieldViolations: []*errdetails.BadRequest_FieldViolation{
		// 			{
		// 				Field:       "account_number",
		// 				Description: "Invalid account number",
		// 			},
		// 		},
		// 	})

		// 	return s.Err()
		// } else if err != nil && accountUuid != uuid.Nil {
		// 	s := status.New(codes.InvalidArgument, err.Error())
		// 	s, _ = s.WithDetails(&errdetails.BadRequest{
		// 		FieldViolations: []*errdetails.BadRequest_FieldViolation{
		// 			{
		// 				Field:       "amount",
		// 				Description: fmt.Sprintf("Requested amount %v exceed available balance", req.Amount),
		// 			},
		// 		},
		// 	})

		// 	return s.Err()
		// }

		if err != nil {
			log.Println("Error while creating transaction :", err)
		}

		err = a.bankService.CalculateTransactionSummary(&tsum, tcur)

		if err != nil {
			return err
		}
	}
}
