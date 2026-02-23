package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	assessment, err := a.llm.Analyze(ctx, txData, userProfile)
	if err != nil {
		return domain.RiskAssessment{IsBlocked: false, Reason: "fail-safe: ai error"}, nil
	}

	amount := extractAmount(txData)
	maxTx := extractMaxTx(userProfile)

	if amount < 500 && assessment.IsBlocked && strings.Contains(strings.ToLower(assessment.Reason), "amount") {
		assessment.IsBlocked = false
		assessment.Reason = "[Low Value Pass] " + assessment.Reason
		return assessment, nil
	}

	if maxTx > 0 && amount > maxTx*2 && amount > 500 {
		assessment.IsBlocked = true
		assessment.Reason = fmt.Sprintf("[Heuristic Block] Transaction amount (%.2f) exceeds historical max (%.2f) significantly", amount, maxTx)
		return assessment, nil
	}

	if amount > 5000 && assessment.ConfidenceScore > 85 && !assessment.IsBlocked {
		assessment.Reason = "[High Value Approved] " + assessment.Reason
		return assessment, nil
	}

	if amount > 10000 && assessment.ConfidenceScore <= 75 {
		assessment.IsBlocked = false
		assessment.Reason = "[PENDING REVIEW] High value anomaly detected: " + assessment.Reason
		return assessment, nil
	}

	switch {
	case assessment.ConfidenceScore <= 30:
		assessment.IsBlocked = false
		assessment.Reason = "[Low Confidence] " + assessment.Reason
	case assessment.ConfidenceScore <= 75:
		if assessment.IsBlocked {
			assessment.IsBlocked = false
			assessment.Reason = "[PENDING REVIEW] " + assessment.Reason
		}
	}

	return assessment, nil
}

func extractAmount(data string) float64 {
	re := regexp.MustCompile(`Amount:?\s*([0-9]+(\.[0-9]+)?)`)
	matches := re.FindStringSubmatch(data)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}

func extractMaxTx(profile string) float64 {
	re := regexp.MustCompile(`MaxTx:?\s*([0-9]+(\.[0-9]+)?)`)
	matches := re.FindStringSubmatch(profile)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}
