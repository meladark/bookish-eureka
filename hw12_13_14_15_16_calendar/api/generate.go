//go:generate protoc --proto_path=. --go_out=../internal/pb --go-grpc_out=../internal/pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative Event.proto EventService.proto
package api
