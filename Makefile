.PHONY: proto test container dapr
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/events.proto

test:
	go test -v ./...

run:
	go run cmd/server.go

dapr_run:
	dapr run --app-id=live-audio-mixer --app-port 50051  --resources-path ./dapr/components -- go run cmd/server.go

dapr:
	dapr run --app-id=live-audio-mixer --app-port 50051  --resources-path ./dapr/components

container:
	docker build -t live-audio-mixer:latest .