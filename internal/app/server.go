package app

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	delivery "github.com/tokyosplif/ai-risk-engine/internal/delivery/grpc"
	"github.com/tokyosplif/ai-risk-engine/internal/infrastructure/llm"
	"github.com/tokyosplif/ai-risk-engine/internal/usecase"
	"github.com/tokyosplif/ai-risk-engine/pkg/pb"
	"google.golang.org/grpc"
)

func RunServer() error {
	apiKey := os.Getenv("GROQ_API_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("PORT env variable is required")
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	groq := llm.NewGroqClient(apiKey)
	analyzer := usecase.NewAnalyzer(groq)
	handler := delivery.NewRiskHandler(analyzer)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRiskEngineServiceServer(grpcServer, handler)

	slog.Info("🧠 AI Risk Engine gRPC server is running", "port", port)
	return grpcServer.Serve(lis)
}
