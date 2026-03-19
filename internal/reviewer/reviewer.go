package reviewer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Pthahnix/Vigilon/internal/config"
)

type ReviewResult struct {
	Priority string `json:"priority"`
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
}

type Reviewer struct {
	Config *config.LLMConfig
}

func New(cfg *config.LLMConfig) *Reviewer {
	return &Reviewer{Config: cfg}
}

const systemPrompt = `You are Vigilon, an AI GPU resource manager for a shared lab server with 3x RTX 5090.
Evaluate the user's GPU resource request and assign a priority level.

Priority levels:
- P0 (Normal): max 1 GPU. For daily debugging, small experiments, inference.
- P1 (Boost): max 2 GPUs. For project sprints, mid-term experiments, model training.
- P2 (Urgent): max 3 GPUs (all). For paper deadlines, urgent experiment supplements.

Respond in JSON only:
{"priority": "P0|P1|P2", "duration": "Nh (hours)", "reason": "brief explanation"}`

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func (r *Reviewer) Review(application string) (*ReviewResult, error) {
	if r.Config.EnvFile != "" {
		loadEnvFile(r.Config.EnvFile)
	}

	apiKey := os.Getenv(r.Config.APIKeyEnv)
	baseURL := os.Getenv(r.Config.BaseURLEnv)
	model := os.Getenv(r.Config.ModelEnv)

	if apiKey == "" {
		return nil, fmt.Errorf("env %s not set", r.Config.APIKeyEnv)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("env %s not set", r.Config.BaseURLEnv)
	}
	if model == "" {
		return nil, fmt.Errorf("env %s not set", r.Config.ModelEnv)
	}

	return r.callOpenAICompat(baseURL, apiKey, model, application)
}

func (r *Reviewer) callOpenAICompat(baseURL, apiKey, model, application string) (*ReviewResult, error) {
	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"

	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": application},
		},
		"max_tokens": 256,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, respData)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respData, &apiResp)
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	text := apiResp.Choices[0].Message.Content
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}") + 1
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON in response: %s", text)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(text[start:end]), &result); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}
	return &result, nil
}
