package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/tokyosplif/ai-risk-engine/internal/domain"
)

type MockLLMClient struct {
	Response domain.RiskAssessment
	Err      error
}

func (m *MockLLMClient) Analyze(ctx context.Context, txData string, userProfile string) (domain.RiskAssessment, error) {
	return m.Response, m.Err
}

func TestProcessAnalysis_LowValuePass(t *testing.T) {
	mockAI := &MockLLMClient{
		Response: domain.RiskAssessment{
			IsBlocked:       true,
			Reason:          "Amount is too high for typical user behavior",
			ConfidenceScore: 90,
		},
	}

	analyzer := NewAnalyzer(mockAI)

	txData := "Amount: 150.0, Merchant: Starbucks, Location: Kyiv"
	userProfile := "MaxTx: 1000.0"

	result, err := analyzer.ProcessAnalysis(context.Background(), txData, userProfile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsBlocked {
		t.Errorf("Expected IsBlocked to be false due to Low Value Pass heuristic, but it's true")
	}

	if !strings.Contains(result.Reason, "[Low Value Pass]") {
		t.Errorf("Expected reason to contain '[Low Value Pass]', got: %s", result.Reason)
	}
}

func TestProcessAnalysis_HeuristicBlock(t *testing.T) {
	mockAI := &MockLLMClient{
		Response: domain.RiskAssessment{
			IsBlocked:       false,
			Reason:          "Normal transaction",
			ConfidenceScore: 95,
		},
	}

	analyzer := NewAnalyzer(mockAI)

	txData := "Amount: 2500.0, Merchant: Apple Store"
	userProfile := "MaxTx: 500.0"

	result, _ := analyzer.ProcessAnalysis(context.Background(), txData, userProfile)

	if !result.IsBlocked {
		t.Errorf("Expected IsBlocked to be true due to Heuristic Block, but it's false")
	}

	if !strings.Contains(result.Reason, "[Heuristic Block]") {
		t.Errorf("Expected reason to contain '[Heuristic Block]', got: %s", result.Reason)
	}
}
