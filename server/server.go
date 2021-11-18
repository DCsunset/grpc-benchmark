package main

import (
	"context"
	"flag"
	"grpc-benchmark/api"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
}

func (s *Server) Call(ctx context.Context, in *api.Request) (*api.Response, error) {
	resp := &api.Response{
		Data: "Hello, " + in.Data,
	}
	return resp, nil
}

func (s *Server) CallStream(stream api.API_CallStreamServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			// normal end
			if err == io.EOF {
				return nil
			}
			return err
		}

		resp := &api.Response{
			Data: "Hello, " + in.Data,
		}
		stream.Send(resp)
	}
}

func main() {
	address := flag.String("address", ":8900", "address to listen at")
	flag.Parse()

	lis, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Server listened at %v", *address)

	server := Server{}
	grpcServer := grpc.NewServer()
	api.RegisterAPIServer(grpcServer, &server)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
