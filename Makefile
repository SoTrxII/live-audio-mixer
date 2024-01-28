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
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --app-port 50303  --resources-path ./.dapr/resources -- go run cmd/server.go

dapr:
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --app-port 50303 --dapr-grpc-port=50051  --resources-path ./.dapr/resources

# Dapr without waiting on the server, used for testing
dapr_test:
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --dapr-grpc-port=50008  --resources-path ./.dapr/resources

# Launch all test with Dapr sidecar enabled
test_with_dapr:
	dapr run --app-id=live-audio-mixer  --resources-path ./.dapr/resources -- go test -v ./... -covermode=atomic -coverprofile=coverage.out

container:
	docker build -t live-audio-mixer:latest .

container_run: container
	dapr run --app-id=live-audio-mixer --dapr-http-max-request-size 16 --app-port 50051  --resources-path ./.dapr/resources -- docker run -p 50051:50051 live-audio-mixer:latest


#Memo
# Convert url to flac 48k
#ffmpeg -i "<url>" -vn -ac 2 -ar 48000 -acodec flac -t 15 <name>.flac