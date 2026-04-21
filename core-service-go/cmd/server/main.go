package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"core-service-go/internal/grpcserver"
	"core-service-go/internal/infra"
	pb "core-service-go/proto"
	"google.golang.org/grpc"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	maxRecent, _ := strconv.ParseInt(getenv("MAX_CACHE_RECENT", "3000"), 10, 64)
	store, err := infra.NewStore(ctx,
		getenv("VALKEY_ADDR", "valkey:6379"),
		getenv("MONGO_URI", "mongodb://mongo:27017"),
		getenv("MONGO_DB", "social"), getenv("MONGO_COLLECTION", "posts"),
		getenv("KAFKA_BROKER", "redpanda:9092"), getenv("KAFKA_TOPIC", "post-events"), maxRecent)
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterPostServiceServer(grpcServer, &grpcserver.Server{Store: store})
	log.Println("core service listening on :9090")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
