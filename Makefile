generate-proto:
	protoc --proto_path=./contracts/proto --go_out=. --go-grpc_out=. --grpc-gateway_out . ./contracts/proto/rooms/rooms.proto
