package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sashabaranov/go-openai"
	"github.com/tokyosplif/ai-risk-engine/internal/config"
	"github.com/tokyosplif/ai-risk-engine/internal/domain"
	"github.com/tokyosplif/ai-risk-engine/pkg/closer"
)

const (
	DefaultLLMTimeout = 15 * time.Second
	maxIdleConns      = 50
	idleConnTimeout   = 30 * time.Second
	llmTemperature    = 0.1
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

	openaiCfg.HTTPClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConns,
			MaxIdleConnsPerHost: maxIdleConns,
			IdleConnTimeout:     idleConnTimeout,
			DisableCompression:  true,
		},
	}

	gc := &GroqClient{
		client:  openai.NewClientWithConfig(openaiCfg),
		model:   cfg.Model,
		prompts: make(map[string]PromptConfig),
	}

	gc.loadPrompts(promptsPath)

	go gc.WatchPrompts(context.Background(), promptsPath)

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

func (g *GroqClient) WatchPrompts(ctx context.Context, path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create watcher", "err", err)
		return
	}

	defer closer.Close(watcher, "fsnotify watcher")

	events := make(chan fsnotify.Event)
	errs := make(chan error)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				events <- event
			case e, ok := <-watcher.Errors:
				if !ok {
					return
				}
				errs <- e
			}
		}
	}()

	if err := watcher.Add(path); err != nil {
		slog.Error("failed to add file to watcher", "path", path, "err", err)
		return
	}

	for {
		select {
		case event := <-events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				slog.Info("Detected change in prompts file, reloading...", "path", path)
				g.loadPrompts(path)
			}
		case e := <-errs:
			slog.Error("watcher error", "err", e)
		case <-ctx.Done():
			slog.Debug("stopping prompts watcher", "path", path)
			return
		}
	}
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
	ctx, cancel := context.WithTimeout(ctx, DefaultLLMTimeout)
	defer cancel()

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
		Temperature: llmTemperature,
	})

	if err != nil {
		slog.Error("ai provider request failed", "err", err)
		return domain.RiskAssessment{}, err
	}

	if len(resp.Choices) == 0 {
		slog.Error("ai provider returned empty choices")
		return domain.RiskAssessment{}, fmt.Errorf("empty choices from ai provider")
	}

	content := resp.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		slog.Error("ai provider returned empty content in choice")
		return domain.RiskAssessment{}, fmt.Errorf("empty content from ai provider")
	}

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
