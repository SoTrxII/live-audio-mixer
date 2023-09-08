package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	object_storage "live-audio-mixer/internal/object-storage"
	pb "live-audio-mixer/proto"
	records_holder "live-audio-mixer/services/records-holder"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
)

const (
	DEFAULT_DAPR_PORT  = 50001
	DEFAULT_STORE_NAME = "object-store"
	DEFAULT_STORE_B64  = true
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedEventStreamServer
	service *records_holder.RecordsHolder
}

func (s *server) Start(ctx context.Context, req *pb.RecordRequest) (*pb.RecordReply, error) {

	err := s.service.Record(req.Id)
	if err != nil {
		slog.Info(fmt.Sprintf(`[Server] :: Couldn't start req "%s": %v`, req.Id, err))
		return nil, err
	}
	slog.Info(fmt.Sprintf(`[Server] :: A new record with id "%s" has started`, req.Id))

	return &pb.RecordReply{Message: fmt.Sprintf("Recording %s started", req.Id)}, nil
}

func (s *server) Stop(ctx context.Context, req *pb.StopRequest) (*pb.StopReply, error) {
	err := s.service.Stop(req.Id)
	if err != nil {
		slog.Info(fmt.Sprintf(`[Server] :: Couldn't stop req "%s": %v`, req.Id, err))
		return nil, err
	}
	slog.Info(fmt.Sprintf(`[Server] :: Record with id "%s" stopped`, req.Id))
	return &pb.StopReply{Message: fmt.Sprintf("Recording %s stopped", req.Id)}, nil
}

func (s *server) StreamEvents(stream pb.EventStream_StreamEventsServer) error {
	for {
		evt, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.EventReply{Message: "Bye"})
		}
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("[Server] :: Received new event : %v", evt))
		err = s.service.Update(evt)
		// An error should not stop the stream
		if err != nil {
			slog.Error(fmt.Sprintf("[Server] :: error handling evt  : %v, %s", evt, err))
		}
	}
}

func main() {
	flag.Parse()
	daprPort := DEFAULT_DAPR_PORT
	if envPort, err := strconv.ParseInt(os.Getenv("DAPR_GRPC_PORT"), 10, 32); err == nil && envPort != 0 {
		daprPort = int(envPort)
	}
	slog.Info("[Main] :: Dapr port is " + strconv.Itoa(daprPort))

	objStoreName := os.Getenv("OBJECT_STORE_NAME")
	if objStoreName == "" {
		objStoreName = DEFAULT_STORE_NAME
	}
	objStoreUseB64 := DEFAULT_STORE_B64
	if useB64, err := strconv.ParseBool(os.Getenv("OBJECT_STORE_B64")); err == nil {
		objStoreUseB64 = useB64
	}

	// Initialize Dapr and object storage
	daprClient, err := makeDaprClient(daprPort, 100)
	ctx := context.Background()
	store := object_storage.NewObjectStorage(&ctx, daprClient, objStoreName, objStoreUseB64)

	// Strat the gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterEventStreamServer(s, &server{service: records_holder.NewRecordsHolder(store)})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func makeDaprClient(port, maxRequestSizeMB int) (client.Client, error) {
	var opts []grpc.CallOption
	opts = append(opts, grpc.MaxCallRecvMsgSize(maxRequestSizeMB*1024*1024))
	conn, err := grpc.Dial(net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)),
		grpc.WithDefaultCallOptions(opts...), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return client.NewClientWithConnection(conn), nil
}
