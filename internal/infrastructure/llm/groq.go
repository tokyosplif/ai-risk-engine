package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/tokyosplif/ai-risk-engine/internal/domain"
)

type PromptConfig struct {
	SystemRole        string   `json:"system_role"`
	SecurityProtocols []string `json:"security_protocols"`
	OutputFormat      string   `json:"output_format"`
}

type GroqClient struct {
	client  *openai.Client
	model   string
	prompts map[string]PromptConfig
}

func NewGroqClient(apiKey string) *GroqClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	gc := &GroqClient{
		client:  openai.NewClientWithConfig(config),
		model:   "llama-3.3-70b-versatile",
		prompts: make(map[string]PromptConfig),
	}

	gc.loadPrompts()
	return gc
}

func (g *GroqClient) loadPrompts() {
	data, err := os.ReadFile("prompts.json")
	if err != nil {
		slog.Error("failed to load prompts.json", "err", err)
		return
	}

	if err := json.Unmarshal(data, &g.prompts); err != nil {
		slog.Error("failed to parse prompts.json", "err", err)
		return
	}
	slog.Info("ai prompts loaded", "count", len(g.prompts))
}

func (g *GroqClient) buildPrompt(version string, userProfile string) string {
	p, ok := g.prompts[version]
	if !ok {
		return "Analyze for fraud. Return JSON."
	}

	protocols := strings.Join(p.SecurityProtocols, "\n- ")
	return fmt.Sprintf("%s\n\nUSER CONTEXT: %s\n\nPROTOCOLS:\n- %s\n\n%s",
		p.SystemRole, userProfile, protocols, p.OutputFormat)
}

func (g *GroqClient) Analyze(ctx context.Context, txData string, userProfile string) (domain.RiskAssessment, error) {
	systemPrompt := g.buildPrompt("antifraud_v1", userProfile)

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("Analyze Transaction: %s", txData)},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.1,
	})

	if err != nil {
		slog.Error("ai provider request failed", "err", err)
		return domain.RiskAssessment{}, err
	}

	content := resp.Choices[0].Message.Content
	var res domain.RiskAssessment

	if err := json.Unmarshal([]byte(content), &res); err != nil {
		slog.Error("ai response parse failed", "content", content, "err", err)
		return domain.RiskAssessment{}, err
	}

	if !res.IsBlocked {
		reasonLower := strings.ToLower(res.Reason)
		suspicious := []string{"exceeds", "high-risk", "anomaly", "binance", "suspicious", "nigeria", "singapore"}
		for _, word := range suspicious {
			if strings.Contains(reasonLower, word) {
				res.IsBlocked = true
				res.Reason = "[Heuristic Block] " + res.Reason
				slog.Warn("heuristic block triggered", "pattern", word)
				break
			}
		}
	}

	slog.Info("risk analysis complete",
		"blocked", res.IsBlocked,
		"reason", res.Reason,
	)

	return res, nil
}
