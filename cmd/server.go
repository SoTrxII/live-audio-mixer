package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"io"
	pb "live-audio-mixer/proto"
	records_holder "live-audio-mixer/services/records-holder"
	"log"
	"log/slog"
	"net"
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterEventStreamServer(s, &server{service: records_holder.NewRecordsHolder()})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
