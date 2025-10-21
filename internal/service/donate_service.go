package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
)

// DonateService 投喂服务
type DonateService struct {
	donateRepo      *repository.DonateRepository
	keyRepo         *repository.KeyRepository
	userRepo        *repository.UserRepository
	adminConfigRepo *repository.AdminConfigRepository
	kyxClient       *KyxClient
	cacheService    *CacheService
	httpClient      *http.Client
	logger          *logrus.Logger
}

// NewDonateService 创建投喂服务
func NewDonateService(
	donateRepo *repository.DonateRepository,
	keyRepo *repository.KeyRepository,
	userRepo *repository.UserRepository,
	adminConfigRepo *repository.AdminConfigRepository,
	kyxClient *KyxClient,
	cacheService *CacheService,
	logger *logrus.Logger,
) *DonateService {
	return &DonateService{
		donateRepo:      donateRepo,
		keyRepo:         keyRepo,
		userRepo:        userRepo,
		adminConfigRepo: adminConfigRepo,
		kyxClient:       kyxClient,
		cacheService:    cacheService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// DonateKeys 投喂Keys
func (s *DonateService) DonateKeys(ctx context.Context, linuxDoID string, keys []string) (*model.DonateResponse, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("User not found for donate")
		return nil, fmt.Errorf("user not found")
	}

	// 检查是否已绑定账号
	if user.KyxUserID == 0 {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("Attempt to donate without bound account")
		return nil, fmt.Errorf("account not bound, please bind first")
	}

	// 检查投喂限制（每天最多投喂次数）
	donateCount, err := s.cacheService.GetDonateCount(ctx, linuxDoID)
	if err == nil && donateCount >= 10 { // 每天最多10次
		s.logger.WithField("linux_do_id", linuxDoID).Warn("Donate limit exceeded")
		return nil, fmt.Errorf("daily donate limit exceeded (max 10 times per day)")
	}

	// 验证Keys
	validationResults, validKeys := s.ValidateKeys(ctx, keys)

	if len(validKeys) == 0 {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("No valid keys to donate")
		return &model.DonateResponse{
			ValidKeys:        0,
			AlreadyExists:    len(keys),
			DuplicateRemoved: 0,
			QuotaAdded:       0,
			Results:          validationResults,
		}, nil
	}

	// 推送Keys到公益站
	pushResult := s.PushKeys(ctx, validKeys)

	// 计算总额度
	totalQuota := int64(len(pushResult.SuccessKeys)) * 500000 // 每个Key = $1 = 500000

	// 如果有成功的Key，添加额度
	if len(pushResult.SuccessKeys) > 0 {
		if err := s.kyxClient.AddQuota(ctx, user.KyxUserID, totalQuota); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"linux_do_id": linuxDoID,
				"kyx_user_id": user.KyxUserID,
				"quota":       totalQuota,
			}).Error("Failed to add quota for donated keys")
			// 继续处理，记录错误但不中断
		}
	}

	// 保存已使用的Keys
	usedKeys := make([]*model.UsedKey, 0, len(pushResult.SuccessKeys))
	for _, key := range pushResult.SuccessKeys {
		usedKeys = append(usedKeys, &model.UsedKey{
			KeyHash:   repository.HashKey(key),
			FullKey:   key,
			LinuxDoID: linuxDoID,
			Username:  user.Username,
			UsedAt:    time.Now(),
		})
	}

	if len(usedKeys) > 0 {
		addedCount, err := s.keyRepo.AddBatch(ctx, usedKeys)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to save used keys")
		}
		s.logger.WithField("added_count", addedCount).Debug("Used keys saved")

		// 更新布隆过滤器
		keyHashes := make([]string, 0, len(usedKeys))
		for _, key := range usedKeys {
			keyHashes = append(keyHashes, key.KeyHash)
		}
		_ = s.cacheService.BloomFilterAddBatch(ctx, keyHashes)
	}

	// 创建投喂记录
	pushStatus := "success"
	pushMessage := fmt.Sprintf("Successfully pushed %d keys", len(pushResult.SuccessKeys))
	if len(pushResult.FailedKeys) > 0 {
		pushStatus = "partial"
		pushMessage = fmt.Sprintf("Pushed %d keys, %d failed", len(pushResult.SuccessKeys), len(pushResult.FailedKeys))
	}
	if len(pushResult.SuccessKeys) == 0 {
		pushStatus = "failed"
		pushMessage = "All keys failed to push"
	}

	record := &model.DonateRecord{
		LinuxDoID:       linuxDoID,
		Username:        user.Username,
		KeysCount:       len(pushResult.SuccessKeys),
		TotalQuotaAdded: totalQuota,
		PushStatus:      pushStatus,
		PushMessage:     pushMessage,
		FailedKeys:      pushResult.FailedKeys,
	}

	if err := s.donateRepo.Create(ctx, record); err != nil {
		s.logger.WithError(err).Error("Failed to create donate record")
		// 不返回错误，因为额度已经添加
	}

	// 增加投喂计数
	_, _ = s.cacheService.IncrDonateCount(ctx, linuxDoID)

	// 清除用户额度缓存
	_ = s.cacheService.ClearUserQuota(ctx, linuxDoID)

	s.logger.WithFields(logrus.Fields{
		"linux_do_id":  linuxDoID,
		"username":     user.Username,
		"total_keys":   len(keys),
		"valid_keys":   len(validKeys),
		"success_keys": len(pushResult.SuccessKeys),
		"failed_keys":  len(pushResult.FailedKeys),
		"quota_added":  totalQuota,
		"push_status":  pushStatus,
	}).Info("Keys donated")

	response := &model.DonateResponse{
		ValidKeys:        len(pushResult.SuccessKeys),
		AlreadyExists:    len(keys) - len(validKeys),
		DuplicateRemoved: len(keys) - len(validKeys),
		QuotaAdded:       totalQuota,
		Results:          validationResults,
	}

	return response, nil
}

// ValidateKeys 验证Keys的有效性
func (s *DonateService) ValidateKeys(ctx context.Context, keys []string) ([]model.KeyValidationResult, []string) {
	results := make([]model.KeyValidationResult, 0, len(keys))
	validKeys := make([]string, 0, len(keys))
	seenKeys := make(map[string]bool)

	for _, key := range keys {
		key = strings.TrimSpace(key)

		// 检查Key格式
		if !s.isValidKeyFormat(key) {
			results = append(results, model.KeyValidationResult{
				Key:    key,
				Valid:  false,
				Reason: "Invalid key format",
			})
			continue
		}

		// 检查是否重复（本次提交中）
		if seenKeys[key] {
			results = append(results, model.KeyValidationResult{
				Key:    key,
				Valid:  false,
				Reason: "Duplicate key in this submission",
			})
			continue
		}
		seenKeys[key] = true

		// 检查布隆过滤器（快速检查）
		keyHash := repository.HashKey(key)
		exists, err := s.cacheService.BloomFilterExists(ctx, keyHash)
		if err == nil && exists {
			// 布隆过滤器显示可能存在，需要进一步验证
			dbExists, err := s.keyRepo.Exists(ctx, keyHash)
			if err == nil && dbExists {
				results = append(results, model.KeyValidationResult{
					Key:    key,
					Valid:  false,
					Reason: "Key already used",
				})
				continue
			}
		}

		// Key有效
		validKeys = append(validKeys, key)
		results = append(results, model.KeyValidationResult{
			Key:   key,
			Valid: true,
		})
	}

	return results, validKeys
}

// isValidKeyFormat 检查Key格式是否有效
func (s *DonateService) isValidKeyFormat(key string) bool {
	// Key应该是 sk-xxx 格式，长度在20-200之间
	if len(key) < 20 || len(key) > 200 {
		return false
	}

	if !strings.HasPrefix(key, "sk-") {
		return false
	}

	return true
}

// PushKeysResult 推送Keys的结果
type PushKeysResult struct {
	SuccessKeys []string
	FailedKeys  []string
}

// PushKeys 推送Keys到公益站Keys API
func (s *DonateService) PushKeys(ctx context.Context, keys []string) *PushKeysResult {
	result := &PushKeysResult{
		SuccessKeys: make([]string, 0),
		FailedKeys:  make([]string, 0),
	}

	// 获取Keys API配置
	apiURL, authorization, err := s.adminConfigRepo.GetKeysAPIConfig(ctx)
	if err != nil || apiURL == "" || authorization == "" {
		s.logger.Warn("Keys API not configured, marking all keys as successful (test mode)")
		// 测试模式：如果API未配置，认为所有Key都成功
		result.SuccessKeys = keys
		return result
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"keys": keys,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal keys request")
		result.FailedKeys = keys
		return result
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.WithError(err).Error("Failed to create keys push request")
		result.FailedKeys = keys
		return result
	}

	// 设置请求头
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to push keys to API")
		result.FailedKeys = keys
		return result
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithError(err).Error("Failed to read keys push response")
		result.FailedKeys = keys
		return result
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Keys push failed")
		result.FailedKeys = keys
		return result
	}

	// 解析响应（假设返回成功的Key列表）
	var apiResponse struct {
		Success     bool     `json:"success"`
		SuccessKeys []string `json:"success_keys"`
		FailedKeys  []string `json:"failed_keys"`
		Message     string   `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		s.logger.WithError(err).WithField("response", string(body)).Warn("Failed to parse keys push response")
		// 如果解析失败但状态码是200，假设全部成功
		result.SuccessKeys = keys
		return result
	}

	if apiResponse.Success {
		if len(apiResponse.SuccessKeys) > 0 {
			result.SuccessKeys = apiResponse.SuccessKeys
		} else {
			result.SuccessKeys = keys
		}
		if len(apiResponse.FailedKeys) > 0 {
			result.FailedKeys = apiResponse.FailedKeys
		}
	} else {
		result.FailedKeys = keys
	}

	s.logger.WithFields(logrus.Fields{
		"total_keys":   len(keys),
		"success_keys": len(result.SuccessKeys),
		"failed_keys":  len(result.FailedKeys),
	}).Info("Keys pushed to API")

	return result
}

// GetDonateHistory 获取用户的投喂历史
func (s *DonateService) GetDonateHistory(ctx context.Context, linuxDoID string, page, pageSize int) ([]*model.DonateRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.donateRepo.GetByLinuxDoID(ctx, linuxDoID, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get donate history")
		return nil, 0, fmt.Errorf("failed to get donate history: %w", err)
	}

	total, err := s.donateRepo.CountByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count donate records")
		return nil, 0, fmt.Errorf("failed to count donate records: %w", err)
	}

	return records, total, nil
}

// GetUserDonateStats 获取用户的投喂统计
func (s *DonateService) GetUserDonateStats(ctx context.Context, linuxDoID string) (totalDonates int64, totalKeys int64, totalQuota int64, err error) {
	totalDonates, err = s.donateRepo.CountByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count user donates")
		return 0, 0, 0, fmt.Errorf("failed to count donates: %w", err)
	}

	totalKeys, err = s.donateRepo.GetTotalKeys(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total donated keys")
		return 0, 0, 0, fmt.Errorf("failed to get total keys: %w", err)
	}

	totalQuota, err = s.donateRepo.GetTotalQuota(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total donated quota")
		return 0, 0, 0, fmt.Errorf("failed to get total quota: %w", err)
	}

	return totalDonates, totalKeys, totalQuota, nil
}

// GetTodayStats 获取今日投喂统计
func (s *DonateService) GetTodayStats(ctx context.Context) (count int64, totalKeys int64, totalQuota int64, err error) {
	return s.donateRepo.GetTodayStats(ctx)
}

// ListAllDonates 获取所有投喂记录（管理员）
func (s *DonateService) ListAllDonates(ctx context.Context, page, pageSize int) ([]*model.DonateRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.donateRepo.List(ctx, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list all donates")
		return nil, 0, fmt.Errorf("failed to list donates: %w", err)
	}

	total, err := s.donateRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count all donates")
		return nil, 0, fmt.Errorf("failed to count donates: %w", err)
	}

	return records, total, nil
}

// GetFailedDonates 获取失败的投喂记录
func (s *DonateService) GetFailedDonates(ctx context.Context, page, pageSize int) ([]*model.DonateRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.donateRepo.GetFailedRecords(ctx, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get failed donates")
		return nil, 0, fmt.Errorf("failed to get failed donates: %w", err)
	}

	total, err := s.donateRepo.CountFailed(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count failed donates")
		return nil, 0, fmt.Errorf("failed to count failed donates: %w", err)
	}

	return records, total, nil
}

// GetAllDonateStats 获取所有投喂统计
func (s *DonateService) GetAllDonateStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 今日统计
	todayCount, todayKeys, todayQuota, err := s.GetTodayStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get today donate stats")
	} else {
		stats["today_count"] = todayCount
		stats["today_keys"] = todayKeys
		stats["today_quota"] = todayQuota
		stats["today_quota_usd"] = model.QuotaToDollar(todayQuota)
	}

	// 总统计
	totalCount, err := s.donateRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total donate count")
	} else {
		stats["total_count"] = totalCount
	}

	// 失败记录数
	failedCount, err := s.donateRepo.CountFailed(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get failed donate count")
	} else {
		stats["failed_count"] = failedCount
	}

	// 总Key数
	totalKeys, err := s.keyRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total keys count")
	} else {
		stats["total_keys"] = totalKeys
	}

	return stats, nil
}

// DeleteDonateRecord 删除投喂记录（管理员）
func (s *DonateService) DeleteDonateRecord(ctx context.Context, recordID int) error {
	if err := s.donateRepo.Delete(ctx, recordID); err != nil {
		s.logger.WithError(err).WithField("record_id", recordID).Error("Failed to delete donate record")
		return fmt.Errorf("failed to delete donate record: %w", err)
	}

	s.logger.WithField("record_id", recordID).Info("Donate record deleted")
	return nil
}

// GetRecentDonates 获取最近的投喂记录
func (s *DonateService) GetRecentDonates(ctx context.Context, limit int) ([]*model.DonateRecord, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	records, err := s.donateRepo.List(ctx, limit, 0)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get recent donates")
		return nil, fmt.Errorf("failed to get recent donates: %w", err)
	}

	return records, nil
}

// CheckKeyExists 检查Key是否已被使用
func (s *DonateService) CheckKeyExists(ctx context.Context, key string) (bool, error) {
	key = strings.TrimSpace(key)
	keyHash := repository.HashKey(key)

	// 先检查布隆过滤器
	exists, err := s.cacheService.BloomFilterExists(ctx, keyHash)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to check bloom filter")
	} else if !exists {
		// 布隆过滤器说不存在，则一定不存在
		return false, nil
	}

	// 布隆过滤器可能存在，检查数据库
	return s.keyRepo.Exists(ctx, keyHash)
}
