FROM golang:1.22.12-bookworm AS builder
WORKDIR /src/core-service-go
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
COPY core-service-go/go.mod ./
RUN go mod download
COPY core-service-go .
COPY proto /src/proto
RUN mkdir -p proto && protoc --proto_path=/src/proto --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative /src/proto/post.proto
RUN go build -o /out/core-service ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/core-service /app/core-service
EXPOSE 9090
ENTRYPOINT ["/app/core-service"]
