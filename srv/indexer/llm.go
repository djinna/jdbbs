package indexer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Gateway defaults. The VM's keyless LLM gateway speaks the Anthropic
// Messages API; override with INDEXER_LLM_URL / INDEXER_LLM_MODEL.
const (
	DefaultBaseURL = "https://llm.int.exe.xyz"
	DefaultModel   = "claude-sonnet-4-5"
	// MaxOutputTokens is the hard cap per call. A chapter's entries fit in a
	// few thousand tokens; anything near the cap means the model is writing a
	// concordance and the prompt, not the cap, is what to fix.
	MaxOutputTokens = 8000
	// MaxInputChars bounds a single chunk; longer chapters are split at
	// paragraph boundaries (≈ 4 chars/token → ~30k tokens).
	MaxInputChars = 120000
)

// Usage is the token accounting for one call or a whole run.
type Usage struct {
	Calls        int `json:"calls"`
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheRead    int `json:"cache_read_input_tokens,omitempty"`
	CacheWrite   int `json:"cache_creation_input_tokens,omitempty"`
}

// Add accumulates.
func (u *Usage) Add(o Usage) {
	u.Calls += o.Calls
	u.InputTokens += o.InputTokens
	u.OutputTokens += o.OutputTokens
	u.CacheRead += o.CacheRead
	u.CacheWrite += o.CacheWrite
}

// CostUSD estimates the bill from a per-million price table. Prices are the
// published Anthropic list at the time of writing (2026-09); unknown models
// return 0 and the report says so.
func (u Usage) CostUSD(model string) float64 {
	p, ok := prices[model]
	if !ok {
		return 0
	}
	return float64(u.InputTokens)/1e6*p[0] + float64(u.OutputTokens)/1e6*p[1]
}

var prices = map[string][2]float64{
	"claude-sonnet-4-5": {3, 15},
	"claude-sonnet-4-6": {3, 15},
	"claude-sonnet-5":   {3, 15},
	"claude-haiku-4-5":  {1, 5},
	"claude-opus-4-5":   {5, 25},
	"claude-opus-4-6":   {5, 25},
	"claude-opus-5":     {5, 25},
}

// Client calls the gateway. Zero value works with defaults. Set Replay to a
// directory to run offline: each request is keyed by a hash of its prompt
// and answered from `<key>.json` there; with Record set as well, live answers
// are written to that directory so a later run can replay them (tests use
// srv/indexer/testdata).
type Client struct {
	BaseURL string
	Model   string
	HTTP    *http.Client
	Replay  string // directory of canned responses; "" = live
	Record  string // directory to write live responses into; "" = don't
	Log     func(format string, a ...any)
	// Fake answers every prompt in-process (server tests); when set, no
	// network, replay or record.
	Fake func(system, user string) (string, error)
}

// NewClientFromEnv reads INDEXER_LLM_URL, INDEXER_LLM_MODEL, INDEXER_LLM_REPLAY,
// INDEXER_LLM_RECORD.
func NewClientFromEnv() *Client {
	c := &Client{
		BaseURL: os.Getenv("INDEXER_LLM_URL"),
		Model:   os.Getenv("INDEXER_LLM_MODEL"),
		Replay:  os.Getenv("INDEXER_LLM_REPLAY"),
		Record:  os.Getenv("INDEXER_LLM_RECORD"),
	}
	return c
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return DefaultBaseURL
}

func (c *Client) model() string {
	if c.Model != "" {
		return c.Model
	}
	return DefaultModel
}

func (c *Client) logf(format string, a ...any) {
	if c.Log != nil {
		c.Log(format, a...)
	}
}

type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		CacheRead    int `json:"cache_read_input_tokens"`
		CacheWrite   int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// canned is what Record writes and Replay reads.
type canned struct {
	Model  string `json:"model"`
	System string `json:"system"`
	User   string `json:"user"`
	Text   string `json:"text"`
	Usage  Usage  `json:"usage"`
}

func promptKey(model, system, user string) string {
	h := sha256.Sum256([]byte(model + "\x00" + system + "\x00" + user))
	return hex.EncodeToString(h[:8])
}

// Complete sends one system+user prompt and returns the text of the reply.
// Output is capped at MaxOutputTokens; a reply cut off at the cap is an error
// (the caller must not parse half a JSON document).
func (c *Client) Complete(ctx context.Context, system, user string) (string, Usage, error) {
	model := c.model()
	if c.Fake != nil {
		text, err := c.Fake(system, user)
		return text, Usage{Calls: 1}, err
	}
	key := promptKey(model, system, user)
	// Replay is strict (offline); Record doubles as a cache so a run that
	// failed half-way does not pay again for the chapters it already has.
	for _, dir := range []string{c.Replay, c.Record} {
		if dir == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, key+".json"))
		if err != nil {
			if dir == c.Replay {
				return "", Usage{}, fmt.Errorf("replay: no canned response %s for this prompt (%w)", key, err)
			}
			continue
		}
		var cn canned
		if err := json.Unmarshal(b, &cn); err != nil {
			return "", Usage{}, fmt.Errorf("replay %s: %w", key, err)
		}
		c.logf("replay %s (%d in / %d out)", key, cn.Usage.InputTokens, cn.Usage.OutputTokens)
		return cn.Text, cn.Usage, nil
	}

	body, _ := json.Marshal(messagesRequest{
		Model:     model,
		MaxTokens: MaxOutputTokens,
		System:    system,
		Messages:  []message{{Role: "user", Content: user}},
	})
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Minute}
	}
	var resp messagesResponse
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL()+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return "", Usage{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", "keyless")
		req.Header.Set("anthropic-version", "2023-06-01")
		start := time.Now()
		res, err := hc.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}
		rb, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		res.Body.Close()
		if res.StatusCode >= 500 || res.StatusCode == 429 {
			lastErr = fmt.Errorf("gateway %d: %.200s", res.StatusCode, rb)
			time.Sleep(time.Duration(attempt+1) * 3 * time.Second)
			continue
		}
		resp = messagesResponse{}
		if err := json.Unmarshal(rb, &resp); err != nil {
			return "", Usage{}, fmt.Errorf("gateway: bad JSON (%d): %.200s", res.StatusCode, rb)
		}
		if resp.Error != nil {
			return "", Usage{}, fmt.Errorf("gateway: %s: %s", resp.Error.Type, resp.Error.Message)
		}
		if res.StatusCode != 200 {
			return "", Usage{}, fmt.Errorf("gateway %d: %.200s", res.StatusCode, rb)
		}
		c.logf("%s: %d in / %d out, %s", model, resp.Usage.InputTokens, resp.Usage.OutputTokens, time.Since(start).Round(time.Second))
		lastErr = nil
		break
	}
	if lastErr != nil {
		return "", Usage{}, lastErr
	}
	var text string
	for _, part := range resp.Content {
		if part.Type == "text" {
			text += part.Text
		}
	}
	u := Usage{Calls: 1, InputTokens: resp.Usage.InputTokens, OutputTokens: resp.Usage.OutputTokens,
		CacheRead: resp.Usage.CacheRead, CacheWrite: resp.Usage.CacheWrite}
	if resp.StopReason == "max_tokens" {
		return text, u, errors.New("reply truncated at the output token cap")
	}
	if c.Record != "" {
		if err := os.MkdirAll(c.Record, 0755); err == nil {
			b, _ := json.MarshalIndent(canned{Model: model, System: system, User: user, Text: text, Usage: u}, "", " ")
			_ = os.WriteFile(filepath.Join(c.Record, key+".json"), b, 0644)
		}
	}
	return text, u, nil
}
