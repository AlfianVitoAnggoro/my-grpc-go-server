package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/AlfianVitoAnggoro/my-grpc-go-server/internal/port"
	"github.com/AlfianVitoAnggoro/my-grpc-proto/protogen/go/bank"
	"github.com/AlfianVitoAnggoro/my-grpc-proto/protogen/go/hello"
	resl "github.com/AlfianVitoAnggoro/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
)

type GrpcAdapter struct {
	helloService      port.HelloServicePort
	bankService       port.BankServicePort
	resiliencyService port.ResiliencyServicePort
	grpcPort          int
	server            *grpc.Server
	hello.HelloServiceServer
	bank.BankServiceServer
	resl.ResiliencyServiceServer
}

func NewGrpcAdapter(helloService port.HelloServicePort, bankService port.BankServicePort, resiliencyService port.ResiliencyServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		helloService:      helloService,
		bankService:       bankService,
		resiliencyService: resiliencyService,
		grpcPort:          grpcPort,
	}
}

func (a *GrpcAdapter) Run() {
	var err error
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Server listening on port %v", a.grpcPort)

	grpcServer := grpc.NewServer()
	a.server = grpcServer

	hello.RegisterHelloServiceServer(grpcServer, a)
	bank.RegisterBankServiceServer(grpcServer, a)
	resl.RegisterResiliencyServiceServer(grpcServer, a)

	if err = grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve gRPC on port %d : %v\n", a.grpcPort, err)
	}
}

func (a *GrpcAdapter) Stop() {
	a.server.Stop()
}
