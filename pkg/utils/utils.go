package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ========== 字符串工具 ==========

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	id, _ := GenerateRandomString(32)
	return id
}

// HashSHA256 SHA256哈希
func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + suffix
}

// ========== 密码工具 ==========

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash 验证密码哈希
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ========== 时间工具 ==========

// GetToday 获取今天的日期字符串 (YYYY-MM-DD)
func GetToday() string {
	return time.Now().Format("2006-01-02")
}

// GetTodayStart 获取今天开始时间
func GetTodayStart() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// GetTodayEnd 获取今天结束时间
func GetTodayEnd() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// ========== HTTP工具 ==========

// HTTPGet 发送GET请求
func HTTPGet(url string, headers map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// HTTPPost 发送POST请求
func HTTPPost(url string, headers map[string]string, data interface{}) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	// 设置默认Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 设置其他请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// HTTPPostForm 发送表单POST请求
func HTTPPostForm(url string, headers map[string]string, formData map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 构建表单数据
	form := make([]string, 0, len(formData))
	for key, value := range formData {
		form = append(form, fmt.Sprintf("%s=%s", key, value))
	}
	formBody := strings.Join(form, "&")

	req, err := http.NewRequest("POST", url, strings.NewReader(formBody))
	if err != nil {
		return nil, err
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置其他请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ========== JSON工具 ==========

// MustMarshalJSON JSON序列化（忽略错误）
func MustMarshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// PrettyJSON 格式化JSON
func PrettyJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

// ========== 切片工具 ==========

// UniqueStrings 字符串数组去重
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ContainsString 检查字符串是否在数组中
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveString 从数组中移除字符串
func RemoveString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// SplitIntoChunks 将数组分块
func SplitIntoChunks(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// ========== 数值工具 ==========

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个整数中的较大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个int64中的较小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两个int64中的较大值
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ========== 额度工具 ==========

// QuotaToDollar 额度转美元
func QuotaToDollar(quota int64) float64 {
	return float64(quota) / 500000.0
}

// DollarToQuota 美元转额度
func DollarToQuota(dollar float64) int64 {
	return int64(dollar * 500000)
}

// FormatQuota 格式化额度为美元字符串
func FormatQuota(quota int64) string {
	return fmt.Sprintf("$%.2f", QuotaToDollar(quota))
}

// ========== 错误处理工具 ==========

// DefaultIfError 如果有错误则返回默认值
func DefaultIfError[T any](value T, err error, defaultValue T) T {
	if err != nil {
		return defaultValue
	}
	return value
}

// Must panic if error
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// ========== 指针工具 ==========

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// IntPtr 返回int指针
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr 返回int64指针
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr 返回bool指针
func BoolPtr(b bool) *bool {
	return &b
}

// PtrString 安全获取指针字符串值
func PtrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PtrInt 安全获取指针int值
func PtrInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// PtrInt64 安全获取指针int64值
func PtrInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// PtrBool 安全获取指针bool值
func PtrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
