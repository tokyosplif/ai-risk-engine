package grpc

import (
	"context"
	"fmt"

	"github.com/tokyosplif/ai-risk-engine/internal/usecase"
	"github.com/tokyosplif/ai-risk-engine/pkg/pb"
)

type RiskHandler struct {
	pb.UnimplementedRiskEngineServiceServer
	usecase *usecase.Analyzer
}

func NewRiskHandler(u *usecase.Analyzer) *RiskHandler {
	return &RiskHandler{usecase: u}
}

func (h *RiskHandler) AnalyzeTransaction(ctx context.Context, req *pb.AnalyzeRequest) (*pb.AnalyzeResponse, error) {
	txData := fmt.Sprintf("Amount: %f, Merchant: %s, Location: %s", req.Amount, req.Merchant, req.Location)

	result, err := h.usecase.ProcessAnalysis(ctx, txData, req.UserProfileContext)
	if err != nil {
		return nil, err
	}

	return &pb.AnalyzeResponse{
		IsBlocked: result.IsBlocked,
		Reason:    result.Reason,
		AiPushMsg: result.AIPushMessage,
	}, nil
}
