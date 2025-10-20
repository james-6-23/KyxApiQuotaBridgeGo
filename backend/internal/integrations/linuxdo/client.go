package linuxdo

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
    cfg        config.OAuthConfig
    httpClient *http.Client
}

func NewClient(cfg config.OAuthConfig) *Client {
    return &Client{
        cfg:        cfg,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn   int    `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
    form := url.Values{}
    form.Set("client_id", c.cfg.ClientID)
    form.Set("client_secret", c.cfg.ClientSecret)
    form.Set("code", code)
    form.Set("redirect_uri", c.cfg.RedirectURI)
    form.Set("grant_type", "authorization_code")

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, bytes.NewBufferString(form.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    res, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token exchange failed: %s", res.Status)
    }

    var token TokenResponse
    if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
        return nil, err
    }
    return &token, nil
}

type UserInfo struct {
    ID            int64  `json:"id"`
    Username      string `json:"username"`
    Name          string `json:"name"`
    AvatarTemplate string `json:"avatar_template"`
}

func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.UserInfoURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

    res, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("user info request failed: %s", res.Status)
    }

    var user UserInfo
    if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
        return nil, err
    }
    return &user, nil
}
