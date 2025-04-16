run:
	bin/orders

build:
	go build -o bin/orders cmd/main.go

proto:
	protoc   -I=internal/adapters/inbound/grpc/  --go_out=paths=source_relative:internal/adapters/inbound/grpc   --go-grpc_out=paths=source_relative:internal/adapters/inbound/grpc   internal/adapters/inbound/grpc/orders.proto