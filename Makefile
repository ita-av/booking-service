.PHONY: generate build run clean

# Generate gRPC code from proto files
generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/api/proto/booking.proto

# Build the application
build: generate
	go build -o bin/server cmd/server/main.go

# Run the application
run: build
	./bin/server

# Clean generated files and binaries
clean:
	rm -f bin/server
	rm -f pkg/api/generated/*.go