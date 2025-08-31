package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"slog-test/pkg/otelinit"

	grpc_a "slog-test/cmd/grpc/grpc-a"
	grpc_b "slog-test/cmd/grpc/grpc-b"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var tracer = otel.Tracer(os.Getenv("OTEL_SERVICE_NAME"))

// server is used to implement grpc_a.GreeterServer.
type server struct {
	grpc_a.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *grpc_a.HelloRequest) (*grpc_a.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())

	ctx, span := tracer.Start(ctx, "SayHello")
	defer span.End()

	span.SetAttributes(attribute.String("test", "test_service_say_hello"))

	// add open telemetry
	conn, err := grpc.NewClient("service-b:5005", grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := grpc_b.NewGreeterClient(conn)
	reply, err := c.SayWorld(ctx, &grpc_b.WorldRequest{Name: in.GetName()})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	return &grpc_a.HelloReply{Message: "Hello " + reply.GetMessage() + in.GetName()}, nil
}

func main() {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	log.Println("Starting", serviceName)
	shutdown := otelinit.InitgRPCProvider(serviceName)
	defer shutdown(context.Background())

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// wrap with open telemetry
	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpc_a.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
