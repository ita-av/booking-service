FROM golang:1.24-alpine as builder


# Install build dependencies
RUN apk add --no-cache protobuf-dev gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate gRPC code
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/api/proto/booking.proto

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o booking-service ./cmd/server

# Create a minimal image
FROM alpine:3.16
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/booking-service .

# Expose gRPC port
EXPOSE 50051

# Command to run
CMD ["./booking-service"]