package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tokyosplif/ai-risk-engine/internal/domain"
)

const (
	lowValueThreshold        = 500.0
	heuristicBlockMinAmount  = 500.0
	heuristicBlockMultiplier = 2.0
	highValueThreshold       = 10000.0
	highConfidenceThreshold  = 75
	lowConfidenceThreshold   = 30
)

var (
	amountRegex = regexp.MustCompile(`Amount:?\s*([0-9]+(\.[0-9]+)?)`)
	maxTxRegex  = regexp.MustCompile(`MaxTx:?\s*([0-9]+(\.[0-9]+)?)`)
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
		return domain.RiskAssessment{IsBlocked: false, Reason: "fail-safe: ai service unavailable"}, nil
	}

	amount := extractAmount(txData)
	maxTx := extractMaxTx(userProfile)

	if amount < lowValueThreshold && assessment.IsBlocked && strings.Contains(strings.ToLower(assessment.Reason), "amount") {
		assessment.IsBlocked = false
		assessment.Reason = addTag(assessment.Reason, "[Low Value Pass]")
		return assessment, nil
	}

	if maxTx > 0 && amount > maxTx*heuristicBlockMultiplier && amount > heuristicBlockMinAmount {
		assessment.IsBlocked = true
		assessment.Reason = addTag(assessment.Reason, fmt.Sprintf("[Heuristic Block] Amount (%.2f) exceeds historical max (%.2f).", amount, maxTx))
		return assessment, nil
	}

	if amount > highValueThreshold && assessment.ConfidenceScore <= highConfidenceThreshold {
		assessment.IsBlocked = false
		assessment.Reason = addTag(assessment.Reason, "[PENDING REVIEW]")
		return assessment, nil
	}

	switch {
	case assessment.ConfidenceScore <= lowConfidenceThreshold:
		assessment.IsBlocked = false
		assessment.Reason = addTag(assessment.Reason, "[Low Confidence Ignore]")
	case assessment.ConfidenceScore <= highConfidenceThreshold:
		if assessment.IsBlocked {
			assessment.IsBlocked = false
			assessment.Reason = addTag(assessment.Reason, "[PENDING REVIEW]")
		}
	}

	return assessment, nil
}

func addTag(reason, tag string) string {
	if strings.Contains(reason, tag) {
		return reason
	}
	return tag + " " + reason
}

func extractAmount(data string) float64 {
	matches := amountRegex.FindStringSubmatch(data)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}

func extractMaxTx(profile string) float64 {
	matches := maxTxRegex.FindStringSubmatch(profile)
	if len(matches) > 1 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		return val
	}
	return 0
}
