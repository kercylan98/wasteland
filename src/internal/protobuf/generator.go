package protobuf

//go:generate protoc --go_out=./protobuf --go_opt=paths=source_relative --go-grpc_out=./protobuf --go-grpc_opt=paths=source_relative service.proto
