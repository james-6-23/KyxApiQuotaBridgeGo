package kyx

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
)

type Client struct {
    httpClient *http.Client
    cfg        config.KYXConfig
}

func NewClient(cfg config.KYXConfig) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: 15 * time.Second},
        cfg:        cfg,
    }
}

type SearchResponse struct {
    Success bool `json:"success"`
    Message string `json:"message"`
    Data struct {
        Items []User `json:"items"`
        Total int `json:"total"`
    } `json:"data"`
}

type User struct {
    ID         int64  `json:"id"`
    Username   string `json:"username"`
    DisplayName string `json:"display_name"`
    LinuxDoID  string `json:"linux_do_id"`
    Quota      int64  `json:"quota"`
    UsedQuota  int64  `json:"used_quota"`
    Group      string `json:"group"`
}

func (c *Client) SearchUser(ctx context.Context, username string, page, pageSize int) (*SearchResponse, error) {
    endpoint, err := url.Parse(fmt.Sprintf("%s/api/user/search", c.cfg.APIBaseURL))
    if err != nil {
        return nil, err
    }
    q := endpoint.Query()
    q.Set("keyword", username)
    q.Set("p", fmt.Sprintf("%d", page))
    q.Set("page_size", fmt.Sprintf("%d", pageSize))
    endpoint.RawQuery = q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.cfg.Session))
    req.Header.Set("new-api-user", c.cfg.NewAPIUser)

    res, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var response SearchResponse
    if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
        return nil, err
    }
    return &response, nil
}

type UpdateQuotaRequest struct {
    ID       int64  `json:"id"`
    Quota    int64  `json:"quota"`
    Username string `json:"username"`
    Group    string `json:"group"`
}

type UpdateQuotaResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func (c *Client) UpdateQuota(ctx context.Context, req UpdateQuotaRequest) (*UpdateQuotaResponse, error) {
    endpoint := fmt.Sprintf("%s/api/user/", c.cfg.APIBaseURL)
    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Cookie", fmt.Sprintf("session=%s", c.cfg.Session))
    httpReq.Header.Set("new-api-user", c.cfg.NewAPIUser)

    res, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var response UpdateQuotaResponse
    if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
        return nil, err
    }
    return &response, nil
}
