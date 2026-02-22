package usecase

import (
	"context"
	"log/slog"

	"github.com/tokyosplif/ai-risk-engine/internal/domain"
)

type LLMClient interface {
	Analyze(ctx context.Context, txData string, userProfile string) (domain.RiskAssessment, error)
}

type Analyzer struct {
	llm LLMClient
}

func NewAnalyzer(llm LLMClient) *Analyzer {
	return &Analyzer{llm: llm}
}

func (a *Analyzer) ProcessAnalysis(ctx context.Context, txData, userProfile string) (domain.RiskAssessment, error) {
	if txData == "" || txData == "{}" {
		return domain.RiskAssessment{
			IsBlocked: false,
			Reason:    "Empty transaction data, skipped",
		}, nil
	}

	assessment, err := a.llm.Analyze(ctx, txData, userProfile)
	if err != nil {
		slog.Error("LLM analysis failed, applying fail-safe (allow)", "error", err)
		return domain.RiskAssessment{
			IsBlocked:     false,
			Reason:        "AI Engine error, fail-safe applied",
			AIPushMessage: "System maintenance in progress",
		}, nil
	}

	if assessment.IsBlocked {
		// Record a warning about blocked transactions. Avoid including full user profile in logs.
		slog.Warn("Transaction blocked by AI analysis", "reason", assessment.Reason)
	}

	return assessment, nil
}
