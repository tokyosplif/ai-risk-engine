package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/tokyosplif/ai-risk-engine/internal/config"
	delivery "github.com/tokyosplif/ai-risk-engine/internal/delivery/grpc"
	"github.com/tokyosplif/ai-risk-engine/internal/infrastructure/llm"
	"github.com/tokyosplif/ai-risk-engine/internal/usecase"
	"github.com/tokyosplif/ai-risk-engine/pkg/pb"
	"google.golang.org/grpc"
)

func RunServer(cfg *config.Config) error {
	groq := llm.NewGroqClient(cfg.Groq, cfg.PromptsPath)

	analyzer := usecase.NewAnalyzer(groq)

	handler := delivery.NewRiskHandler(analyzer)

	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRiskEngineServiceServer(grpcServer, handler)

	slog.Info("AI Risk Engine gRPC server is running", "port", cfg.Port)
	return grpcServer.Serve(lis)
}
