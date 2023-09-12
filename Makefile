.PHONY: proto test container dapr
ifeq ($(OS),Windows_NT)
    DETACHED_EXEC = cmd.exe /c start /b /min
else
    DETACHED_EXEC = nohup
endif

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/events.proto

test:
	go test -v ./...

run:
	go run cmd/server.go

dapr_run:
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --app-port 50051  --resources-path ./dapr/components -- go run cmd/server.go

dapr:
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --app-port 50051 --dapr-grpc-port=50008  --resources-path ./dapr/components

# Dapr without waiting on the server, used for testing
dapr_test:
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --dapr-grpc-port=50008  --resources-path ./dapr/components

# Launch all test with Dapr sidecar enabled
test_with_dapr:
	dapr run --app-id=live-audio-mixer  --resources-path ./dapr/components -- go test -v ./... -covermode=atomic -coverprofile=coverage.out

container:
	docker build -t live-audio-mixer:latest .