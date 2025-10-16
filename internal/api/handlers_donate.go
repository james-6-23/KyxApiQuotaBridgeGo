package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kyx-api-quota-bridge/internal/models"
)

// handleDonateValidate 投喂 Keys
func (s *Server) handleDonateValidate(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		sendError(c, http.StatusUnauthorized, "未登录")
		return
	}

	session, err := s.db.GetSession(sessionID)
	if err != nil || session == nil {
		sendError(c, http.StatusUnauthorized, "会话无效")
		return
	}

	user, err := s.db.GetUser(session.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	if user == nil {
		sendError(c, http.StatusBadRequest, "未绑定账号")
		return
	}

	var req struct {
		Keys []string `json:"keys"`
	}

	if err := c.BindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if len(req.Keys) == 0 {
		sendError(c, http.StatusBadRequest, "Keys 不能为空")
		return
	}

	originalCount := len(req.Keys)

	// 去重
	uniqueKeys := make(map[string]bool)
	for _, key := range req.Keys {
		uniqueKeys[key] = true
	}
	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}
	duplicateCount := originalCount - len(keys)

	// 检查数据库已存在的 Keys
	alreadyExistsKeys := []string{}
	keysToValidate := []string{}

	for _, key := range keys {
		exists, _ := s.db.IsKeyUsed(key)
		if exists {
			alreadyExistsKeys = append(alreadyExistsKeys, key)
		} else {
			keysToValidate = append(keysToValidate, key)
		}
	}

	// 并发验证 Keys
	validKeys := []string{}
	invalidKeys := []string{}
	results := []map[string]interface{}{}

	// 添加已存在的 Keys 到结果
	for _, key := range alreadyExistsKeys {
		results = append(results, map[string]interface{}{
			"key":    key[:10] + "...",
			"valid":  false,
			"reason": "数据库已存在",
		})
	}

	// 并发验证（10个并发）
	batchSize := 10
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < len(keysToValidate); i += batchSize {
		end := i + batchSize
		if end > len(keysToValidate) {
			end = len(keysToValidate)
		}

		batch := keysToValidate[i:end]
		for _, key := range batch {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				isValid := validateModelScopeKey(k, s.config.ModelScopeAPIBase)

				mu.Lock()
				defer mu.Unlock()

				if isValid {
					validKeys = append(validKeys, k)
					s.db.MarkKeyUsed(k, user.LinuxDoID, user.Username)
					results = append(results, map[string]interface{}{
						"key":   k[:10] + "...",
						"valid": true,
					})
				} else {
					invalidKeys = append(invalidKeys, k)
					results = append(results, map[string]interface{}{
						"key":    k[:10] + "...",
						"valid":  false,
						"reason": "无效",
					})
				}
			}(key)
		}
		wg.Wait()
	}

	if len(validKeys) == 0 {
		message := fmt.Sprintf("提交了 %d 个Key，去重后 %d 个，数据库已存在 %d 个，验证后无有效Key",
			originalCount, len(keys), len(alreadyExistsKeys))
		sendError(c, http.StatusBadRequest, message)
		return
	}

	// 更新用户额度
	totalQuotaAdded := int64(len(validKeys)) * s.config.DonateQuotaPerKey
	config, _ := s.db.GetAdminConfig()

	kyxResult, _ := searchKyxUser(user.Username, config.Session, config.NewAPIUser, s.config.KyxAPIBase)
	if kyxResult != nil && kyxResult.Success {
		kyxUser := findExactUser(kyxResult, user.Username)
		if kyxUser != nil {
			newQuota := kyxUser.Quota + totalQuotaAdded
			updateKyxUserQuota(user.KyxUserID, newQuota, config.Session, config.NewAPIUser,
				kyxUser.Username, kyxUser.Group, s.config.KyxAPIBase)
		}
	}

	// 推送 Keys 到分组
	pushStatus := "success"
	pushMessage := "推送成功"
	var failedKeys []string

	if config.KeysAuthorization != "" {
		success, msg, failed := pushKeysToGroup(validKeys, config.KeysAPIURL, config.KeysAuthorization, config.GroupID)
		if !success {
			pushStatus = "failed"
			pushMessage = msg
			failedKeys = failed
		}
	} else {
		pushStatus = "failed"
		pushMessage = "未配置推送授权"
		failedKeys = validKeys
	}

	// 保存投喂记录
	record := &models.DonateRecord{
		LinuxDoID:       user.LinuxDoID,
		Username:        user.Username,
		KeysCount:       len(validKeys),
		TotalQuotaAdded: totalQuotaAdded,
		Timestamp:       time.Now().Unix(),
		PushStatus:      pushStatus,
		PushMessage:     pushMessage,
		FailedKeys:      failedKeys,
	}

	s.db.AddDonateRecord(record)

	// 构建详细消息
	message := ""
	if duplicateCount > 0 {
		message += fmt.Sprintf("已自动去重 %d 个重复Key。", duplicateCount)
	}
	if len(alreadyExistsKeys) > 0 {
		message += fmt.Sprintf("数据库已存在 %d 个Key。", len(alreadyExistsKeys))
	}
	if len(invalidKeys) > 0 {
		message += fmt.Sprintf("发现 %d 个无效Key。", len(invalidKeys))
	}
	message += fmt.Sprintf("成功投喂 %d 个有效Key，已为您增加 ¥%.2f 额度！", len(validKeys), float64(totalQuotaAdded)*0.02)

	sendSuccessWithMessage(c, message, map[string]interface{}{
		"valid_keys":        len(validKeys),
		"already_exists":    len(alreadyExistsKeys),
		"duplicate_removed": duplicateCount,
		"quota_added":       totalQuotaAdded,
		"push_status":       pushStatus,
		"results":           results,
	})
}
