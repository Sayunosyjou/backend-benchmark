FROM golang:1.22.12-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/seeder ./cmd/seeder

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/seeder /app/seeder
ENTRYPOINT ["/app/seeder"]
