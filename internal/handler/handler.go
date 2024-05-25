package handler

import (
	"context"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	pb "github.com/kanataidarov/tinkoff_voicekit/pkg/teams_voicein"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type server struct {
	pb.UnimplementedSpeechToTextServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HwRequest) (*pb.HwResponse, error) {
	return &pb.HwResponse{Message: "H, W!"}, nil
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
