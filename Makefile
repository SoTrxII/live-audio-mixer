.PHONY: proto test container
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/events.proto

test:
	go test -v ./...

run:
	go run cmd/server.go

container:
	docker build -t live-audio-mixer:latest .