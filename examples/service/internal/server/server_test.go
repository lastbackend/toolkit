package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/lastbackend/toolkit/pkg/runtime/controller"

	"google.golang.org/grpc/credentials/insecure"

	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	ptypes "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type ExampleServer struct {
	servicepb.UnimplementedExampleServer
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	runtime, _ := controller.NewRuntime(context.Background(), "test")
	server := NewServer(runtime.Service(), nil)
	servicepb.RegisterExampleServer(srv, server)

	go func() {
		if err := srv.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestExampleServer_HelloWorld(t *testing.T) {
	tests := []struct {
		name    string
		amount  float32
		res     *ptypes.HelloWorldResponse
		errCode codes.Code
		errMsg  string
	}{
		{
			"empty hello world response",
			-1.11,
			nil,
			codes.InvalidArgument,
			fmt.Sprintf("cannot deposit %v", -1.11),
		},
		{
			"spaced hello world response",
			0.00,
			&ptypes.HelloWorldResponse{
				Id:        "1",
				Name:      "",
				Type:      "",
				Data:      nil,
				CreatedAt: 0,
				UpdatedAt: 0,
			},
			codes.OK,
			"",
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := servicepb.NewExampleClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &ptypes.HelloWorldRequest{}

			response, err := client.HelloWorld(ctx, request)
			if response != nil {
				if tt.res == nil {
					t.Error("rsponse: expected nil. received: ", response)
				} else {
					if response.Id != tt.res.Id {
						t.Error("response: expected", tt.res.Id, "received", response.Id)
					}
				}
			}

			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.errCode {
						t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					}
					if er.Message() != tt.errMsg {
						t.Error("error message: expected", tt.errMsg, "received", er.Message())
					}
				}
			}
		})
	}
}
