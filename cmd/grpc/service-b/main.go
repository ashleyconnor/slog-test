package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"slog-test/pkg/otelinit"

	grpc_b "slog-test/cmd/grpc/grpc-b"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var tracer = otel.Tracer(os.Getenv("OTEL_SERVICE_NAME"))

// server is used to implement helloworld.GreeterServer.
type server struct {
	grpc_b.UnimplementedGreeterServer
}

func (s *server) SayWorld(ctx context.Context, in *grpc_b.WorldRequest) (*grpc_b.WorldReply, error) {
	log.Printf("Received: %v", in.GetName())

	// add open telemetry
	_, span := tracer.Start(ctx, "SayWorld")
	defer span.End()

	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "hi!"))

	return &grpc_b.WorldReply{Message: "World " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	log.Println("Starting", serviceName)
	shutdown := otelinit.InitgRPCProvider(serviceName)
	defer shutdown(context.Background())

	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpc_b.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
