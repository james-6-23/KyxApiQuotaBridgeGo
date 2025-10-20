package modelscope

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
)

type Client struct {
    cfg        config.ModelScopeConfig
    httpClient *http.Client
}

func NewClient(cfg config.ModelScopeConfig) *Client {
    return &Client{
        cfg:        cfg,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *Client) ValidateKey(ctx context.Context, apiKey string) (bool, error) {
    payload := map[string]any{
        "model":   "ZhipuAI/GLM-4.6",
        "messages": []map[string]string{{"role": "user", "content": "test"}},
        "max_tokens": 1,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return false, err
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/chat/completions", c.cfg.APIBaseURL), bytes.NewReader(body))
    if err != nil {
        return false, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

    res, err := c.httpClient.Do(req)
    if err != nil {
        return false, err
    }
    defer res.Body.Close()

    if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusTooManyRequests {
        return true, nil
    }
    return false, nil
}
