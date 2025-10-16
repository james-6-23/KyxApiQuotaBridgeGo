package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kyx-api-quota-bridge/internal/models"
)

// searchKyxUser 搜索公益站用户
func searchKyxUser(username, session, newAPIUser, baseURL string) (*models.KyxSearchResult, error) {
	url := fmt.Sprintf("%s/api/user/search?keyword=%s", baseURL, username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", "session="+session)
	req.Header.Set("new-api-user", newAPIUser)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.KyxSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// updateKyxUserQuota 更新公益站用户额度
func updateKyxUserQuota(userID int64, newQuota int64, session, newAPIUser, username, group, baseURL string) error {
	url := fmt.Sprintf("%s/api/user/", baseURL)
	body := map[string]interface{}{
		"id":       userID,
		"quota":    newQuota,
		"username": username,
		"group":    group,
	}

	bodyJSON, _ := json.Marshal(body)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session="+session)
	req.Header.Set("new-api-user", newAPIUser)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf(result.Message)
	}

	return nil
}

// findExactUser 从搜索结果中找到精确匹配的用户
func findExactUser(result *models.KyxSearchResult, username string) *models.KyxUser {
	if !result.Success || len(result.Data.Items) == 0 {
		return nil
	}

	for _, user := range result.Data.Items {
		if user.Username == username {
			return &user
		}
	}

	return nil
}

// validateModelScopeKey 验证 ModelScope Key
func validateModelScopeKey(apiKey, baseURL string) bool {
	body := map[string]interface{}{
		"model":      "ZhipuAI/GLM-4.6",
		"messages":   []map[string]string{{"role": "user", "content": "test"}},
		"max_tokens": 1,
	}

	bodyJSON, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(bodyJSON))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 200 或 429 都表示 key 有效
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusTooManyRequests
}

// pushKeysToGroup 推送 keys 到分组
func pushKeysToGroup(keys []string, apiURL, authorization string, groupID int) (bool, string, []string) {
	keysText := strings.Join(keys, "\n")
	body := map[string]interface{}{
		"group_id":  groupID,
		"keys_text": keysText,
	}

	bodyJSON, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return false, "推送请求失败: " + err.Error(), keys
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authorization)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "推送请求失败: " + err.Error(), keys
	}
	defer resp.Body.Close()

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "解析响应失败", keys
	}

	if !resp.StatusCode == http.StatusOK || !result.Success {
		return false, result.Message, keys
	}

	return true, "推送成功", nil
}

// parseJSON 解析 JSON
func parseJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}
