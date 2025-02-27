package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AlfianVitoAnggoro/my-grpc-proto/protogen/go/hello"
	"google.golang.org/grpc"
)

func (a *GrpcAdapter) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	greet := a.helloService.GenerateHello(req.Name)

	return &hello.HelloResponse{
		Greet: greet,
	}, nil
}

func (a *GrpcAdapter) SayManyHellos(req *hello.HelloRequest, stream hello.HelloService_SayManyHellosServer) error {
	for i := 0; i < 10; i++ {
		greet := a.helloService.GenerateHello(req.Name)

		res := fmt.Sprintf("[%d], %s", i, greet)

		stream.Send(
			&hello.HelloResponse{
				Greet: res,
			},
		)

		time.Sleep(1000 * time.Millisecond)
	}

	return nil

}

func (a *GrpcAdapter) SayHelloToEveryone(stream grpc.ClientStreamingServer[hello.HelloRequest, hello.HelloResponse]) error {
	res := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&hello.HelloResponse{
				Greet: res,
			})
		}

		if err != nil {
			log.Fatalln("Error while reading from client", err)
		}

		greet := a.helloService.GenerateHello(req.Name)

		res += greet + ", "
	}
}
