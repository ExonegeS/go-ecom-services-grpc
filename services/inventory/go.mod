module github.com/ExonegeS/go-ecom-services-grpc/services/inventory

go 1.24.2

replace github.com/ExonegeS/prettyslog => ../pkg/lib/prettyslog

require github.com/ExonegeS/prettyslog v0.0.0-00010101000000-000000000000

require github.com/lib/pq v1.10.9

require (
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
)
