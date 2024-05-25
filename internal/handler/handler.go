package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	pb "github.com/kanataidarov/tinkoff_voicekit/pkg/teams_voicein"
	"google.golang.org/grpc"
	"log"
	"log/slog"
	"net"
)

type server struct {
	pb.UnimplementedSpeechToTextServer
}

func (s *server) Recognize(ctx context.Context, req *pb.SttRequest) (*pb.SttResponse, error) {
	var (
		header *pb.FileHeader
		buf    bytes.Buffer
	)

	header = req.Header
	log.Printf("File name: %v", header.Name)
	if header.Size != nil {
		log.Printf("File size should be: %v", header.Size)
	}
	if data := req.Data; data != nil {
		buf.Write(data)
	}

	log.Printf("Total bytes received: %v", buf.Len())

	return &pb.SttResponse{Message: fmt.Sprintf("Received %v bytes. Thanks!", buf.Len())}, nil
}

func Serve(cfg *config.Config, log *slog.Logger) {
	port := cfg.Grpc.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error(fmt.Sprintf("Failed to listen at %d. %v", port, err))
	}

	srv := grpc.NewServer()
	pb.RegisterSpeechToTextServer(srv, &server{})
	log.Info(fmt.Sprintf("Server listening on %v", lis.Addr()))
	if err := srv.Serve(lis); err != nil {
		log.Error(fmt.Sprintf("Failed to serve at %d. %v", port, err))
	}
}
