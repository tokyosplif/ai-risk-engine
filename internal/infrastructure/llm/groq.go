package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sashabaranov/go-openai"
	"github.com/tokyosplif/ai-risk-engine/internal/config"
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
	mu      sync.RWMutex
	prompts map[string]PromptConfig
}

func NewGroqClient(cfg config.GroqConfig, promptsPath string) *GroqClient {
	openaiCfg := openai.DefaultConfig(cfg.APIKey)
	openaiCfg.BaseURL = cfg.BaseURL

	gc := &GroqClient{
		client:  openai.NewClientWithConfig(openaiCfg),
		model:   cfg.Model,
		prompts: make(map[string]PromptConfig),
	}

	gc.loadPrompts(promptsPath)

	go gc.WatchPrompts(promptsPath)

	return gc
}

func (g *GroqClient) loadPrompts(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to load prompts file", "path", path, "err", err)
		return
	}

	var newPrompts map[string]PromptConfig
	if err := json.Unmarshal(data, &newPrompts); err != nil {
		slog.Error("failed to parse prompts json", "err", err)
		return
	}

	g.mu.Lock()
	g.prompts = newPrompts
	g.mu.Unlock()

	slog.Debug("ai prompts loaded/reloaded", "count", len(newPrompts))
}

func (g *GroqClient) WatchPrompts(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create watcher", "err", err)
		return
	}

	defer func() {
		if err := watcher.Close(); err != nil {
			slog.Error("failed to close watcher", "err", err)
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					slog.Info("Detected change in prompts file, reloading...", "path", path)
					g.loadPrompts(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("watcher error", "err", err)
			}
		}
	}()

	if err := watcher.Add(path); err != nil {
		slog.Error("failed to add file to watcher", "path", path, "err", err)
	}

	select {}
}

func (g *GroqClient) buildPrompt(version string, userProfile string) string {
	g.mu.RLock()
	p, ok := g.prompts[version]
	g.mu.RUnlock()

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

	slog.Debug("risk analysis complete", "blocked", res.IsBlocked, "reason", res.Reason)
	return res, nil
}
