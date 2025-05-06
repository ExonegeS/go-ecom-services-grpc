run:
	bin/statistics

build:
	go build -o bin/statistics cmd/api/main.go

proto:
	protoc   -I=internal/adapters/grpc/  --go_out=paths=source_relative:internal/adapters/grpc   --go-grpc_out=paths=source_relative:internal/adapters/grpc   internal/adapters/grpc/statistics.proto