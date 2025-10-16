package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyx-api-quota-bridge/internal/models"
	_ "modernc.org/sqlite"
)

// DB 数据库包装器
type DB struct {
	db *sql.DB
}

// NewDB 创建新的数据库连接
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &DB{db: db}
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.db.Close()
}

// createTables 创建所有必需的表
func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			linux_do_id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			kyx_user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			linux_do_id TEXT NOT NULL,
			username TEXT,
			name TEXT,
			avatar_url TEXT,
			admin BOOLEAN DEFAULT FALSE,
			expires_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS claim_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			linux_do_id TEXT NOT NULL,
			username TEXT NOT NULL,
			quota_added INTEGER NOT NULL,
			timestamp INTEGER NOT NULL,
			date TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS donate_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			linux_do_id TEXT NOT NULL,
			username TEXT NOT NULL,
			keys_count INTEGER NOT NULL,
			total_quota_added INTEGER NOT NULL,
			timestamp INTEGER NOT NULL,
			push_status TEXT,
			push_message TEXT,
			failed_keys TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS donated_keys (
			key TEXT PRIMARY KEY,
			linux_do_id TEXT NOT NULL,
			username TEXT NOT NULL,
			timestamp INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS admin_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			session TEXT,
			new_api_user TEXT DEFAULT '1',
			claim_quota INTEGER DEFAULT 20000000,
			keys_api_url TEXT DEFAULT 'https://gpt-load.kyx03.de/api/keys/add-async',
			keys_authorization TEXT,
			group_id INTEGER DEFAULT 26,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_records_linux_do_id ON claim_records(linux_do_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_records_date ON claim_records(date)`,
		`CREATE INDEX IF NOT EXISTS idx_donate_records_linux_do_id ON donate_records(linux_do_id)`,
		`CREATE INDEX IF NOT EXISTS idx_donated_keys_linux_do_id ON donated_keys(linux_do_id)`,
	}

	for _, query := range queries {
		if _, err := db.db.Exec(query); err != nil {
			return err
		}
	}

	// 初始化管理员配置
	_, err := db.db.Exec(`INSERT OR IGNORE INTO admin_config (id, updated_at) VALUES (1, ?)`, time.Now().Unix())
	return err
}

// ==================== User Operations ====================

// GetUser 获取用户
func (db *DB) GetUser(linuxDoID string) (*models.User, error) {
	var user models.User
	err := db.db.QueryRow(`
		SELECT linux_do_id, username, kyx_user_id, created_at 
		FROM users WHERE linux_do_id = ?
	`, linuxDoID).Scan(&user.LinuxDoID, &user.Username, &user.KyxUserID, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUser 设置用户
func (db *DB) SetUser(user *models.User) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO users (linux_do_id, username, kyx_user_id, created_at)
		VALUES (?, ?, ?, ?)
	`, user.LinuxDoID, user.Username, user.KyxUserID, user.CreatedAt)
	return err
}

// GetAllUsers 获取所有用户
func (db *DB) GetAllUsers() ([]models.User, error) {
	rows, err := db.db.Query(`SELECT linux_do_id, username, kyx_user_id, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.LinuxDoID, &user.Username, &user.KyxUserID, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// ==================== Session Operations ====================

// SaveSession 保存会话
func (db *DB) SaveSession(session *models.Session) error {
	_, err := db.db.Exec(`
		INSERT OR REPLACE INTO sessions (session_id, linux_do_id, username, name, avatar_url, admin, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, session.SessionID, session.LinuxDoID, session.Username, session.Name, session.AvatarURL, session.Admin, session.ExpiresAt)
	return err
}

// GetSession 获取会话
func (db *DB) GetSession(sessionID string) (*models.Session, error) {
	var session models.Session
	err := db.db.QueryRow(`
		SELECT session_id, linux_do_id, username, name, avatar_url, admin, expires_at
		FROM sessions WHERE session_id = ?
	`, sessionID).Scan(&session.SessionID, &session.LinuxDoID, &session.Username,
		&session.Name, &session.AvatarURL, &session.Admin, &session.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if session.ExpiresAt < time.Now().Unix() {
		db.DeleteSession(sessionID)
		return nil, nil
	}

	return &session, nil
}

// DeleteSession 删除会话
func (db *DB) DeleteSession(sessionID string) error {
	_, err := db.db.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
	return err
}

// ==================== Claim Records ====================

// AddClaimRecord 添加领取记录
func (db *DB) AddClaimRecord(record *models.ClaimRecord) error {
	_, err := db.db.Exec(`
		INSERT INTO claim_records (linux_do_id, username, quota_added, timestamp, date)
		VALUES (?, ?, ?, ?, ?)
	`, record.LinuxDoID, record.Username, record.QuotaAdded, record.Timestamp, record.Date)
	return err
}

// GetClaimToday 获取今天的领取记录
func (db *DB) GetClaimToday(linuxDoID string) (bool, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := db.db.QueryRow(`
		SELECT COUNT(*) FROM claim_records 
		WHERE linux_do_id = ? AND date = ?
	`, linuxDoID, today).Scan(&count)
	return count > 0, err
}

// GetUserClaimRecords 获取用户领取记录
func (db *DB) GetUserClaimRecords(linuxDoID string) ([]models.ClaimRecord, error) {
	rows, err := db.db.Query(`
		SELECT linux_do_id, username, quota_added, timestamp, date
		FROM claim_records WHERE linux_do_id = ?
		ORDER BY timestamp DESC
	`, linuxDoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ClaimRecord
	for rows.Next() {
		var record models.ClaimRecord
		if err := rows.Scan(&record.LinuxDoID, &record.Username, &record.QuotaAdded,
			&record.Timestamp, &record.Date); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// GetAllClaimRecords 获取所有领取记录
func (db *DB) GetAllClaimRecords() ([]models.ClaimRecord, error) {
	rows, err := db.db.Query(`
		SELECT linux_do_id, username, quota_added, timestamp, date
		FROM claim_records ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ClaimRecord
	for rows.Next() {
		var record models.ClaimRecord
		if err := rows.Scan(&record.LinuxDoID, &record.Username, &record.QuotaAdded,
			&record.Timestamp, &record.Date); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// ==================== Donate Records ====================

// AddDonateRecord 添加投喂记录
func (db *DB) AddDonateRecord(record *models.DonateRecord) error {
	failedKeysJSON, _ := json.Marshal(record.FailedKeys)
	_, err := db.db.Exec(`
		INSERT INTO donate_records (linux_do_id, username, keys_count, total_quota_added, 
			timestamp, push_status, push_message, failed_keys)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, record.LinuxDoID, record.Username, record.KeysCount, record.TotalQuotaAdded,
		record.Timestamp, record.PushStatus, record.PushMessage, string(failedKeysJSON))
	return err
}

// GetUserDonateRecords 获取用户投喂记录
func (db *DB) GetUserDonateRecords(linuxDoID string) ([]models.DonateRecord, error) {
	rows, err := db.db.Query(`
		SELECT id, linux_do_id, username, keys_count, total_quota_added, timestamp, 
			push_status, push_message, failed_keys
		FROM donate_records WHERE linux_do_id = ?
		ORDER BY timestamp DESC
	`, linuxDoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDonateRecords(rows)
}

// GetAllDonateRecords 获取所有投喂记录
func (db *DB) GetAllDonateRecords() ([]models.DonateRecord, error) {
	rows, err := db.db.Query(`
		SELECT id, linux_do_id, username, keys_count, total_quota_added, timestamp,
			push_status, push_message, failed_keys
		FROM donate_records ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDonateRecords(rows)
}

// GetDonateRecord 获取单条投喂记录
func (db *DB) GetDonateRecord(linuxDoID string, timestamp int64) (*models.DonateRecord, error) {
	var record models.DonateRecord
	var failedKeysJSON sql.NullString

	err := db.db.QueryRow(`
		SELECT linux_do_id, username, keys_count, total_quota_added, timestamp,
			push_status, push_message, failed_keys
		FROM donate_records WHERE linux_do_id = ? AND timestamp = ?
	`, linuxDoID, timestamp).Scan(&record.LinuxDoID, &record.Username, &record.KeysCount,
		&record.TotalQuotaAdded, &record.Timestamp, &record.PushStatus,
		&record.PushMessage, &failedKeysJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if failedKeysJSON.Valid && failedKeysJSON.String != "" {
		json.Unmarshal([]byte(failedKeysJSON.String), &record.FailedKeys)
	}

	return &record, nil
}

// UpdateDonateRecord 更新投喂记录
func (db *DB) UpdateDonateRecord(linuxDoID string, timestamp int64, record *models.DonateRecord) error {
	failedKeysJSON, _ := json.Marshal(record.FailedKeys)
	_, err := db.db.Exec(`
		UPDATE donate_records 
		SET push_status = ?, push_message = ?, failed_keys = ?
		WHERE linux_do_id = ? AND timestamp = ?
	`, record.PushStatus, record.PushMessage, string(failedKeysJSON), linuxDoID, timestamp)
	return err
}

func scanDonateRecords(rows *sql.Rows) ([]models.DonateRecord, error) {
	var records []models.DonateRecord
	for rows.Next() {
		var record models.DonateRecord
		var id int64
		var failedKeysJSON sql.NullString

		if err := rows.Scan(&id, &record.LinuxDoID, &record.Username, &record.KeysCount,
			&record.TotalQuotaAdded, &record.Timestamp, &record.PushStatus,
			&record.PushMessage, &failedKeysJSON); err != nil {
			return nil, err
		}

		if failedKeysJSON.Valid && failedKeysJSON.String != "" {
			json.Unmarshal([]byte(failedKeysJSON.String), &record.FailedKeys)
		}

		records = append(records, record)
	}
	return records, rows.Err()
}

// ==================== Donated Keys ====================

// IsKeyUsed 检查 Key 是否已使用
func (db *DB) IsKeyUsed(key string) (bool, error) {
	var count int
	err := db.db.QueryRow(`SELECT COUNT(*) FROM donated_keys WHERE key = ?`, key).Scan(&count)
	return count > 0, err
}

// MarkKeyUsed 标记 Key 为已使用
func (db *DB) MarkKeyUsed(key, linuxDoID, username string) error {
	_, err := db.db.Exec(`
		INSERT OR IGNORE INTO donated_keys (key, linux_do_id, username, timestamp)
		VALUES (?, ?, ?, ?)
	`, key, linuxDoID, username, time.Now().Unix())
	return err
}

// GetAllDonatedKeys 获取所有投喂的 Keys
func (db *DB) GetAllDonatedKeys() ([]models.DonatedKey, error) {
	rows, err := db.db.Query(`
		SELECT key, linux_do_id, username, timestamp FROM donated_keys
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.DonatedKey
	for rows.Next() {
		var key models.DonatedKey
		if err := rows.Scan(&key.Key, &key.LinuxDoID, &key.Username, &key.Timestamp); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// DeleteKeys 删除指定的 Keys
func (db *DB) DeleteKeys(keys []string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM donated_keys WHERE key = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, key := range keys {
		if _, err := stmt.Exec(key); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ==================== Admin Config ====================

// GetAdminConfig 获取管理员配置
func (db *DB) GetAdminConfig() (*models.AdminConfig, error) {
	var config models.AdminConfig
	err := db.db.QueryRow(`
		SELECT session, new_api_user, claim_quota, keys_api_url, 
			keys_authorization, group_id, updated_at
		FROM admin_config WHERE id = 1
	`).Scan(&config.Session, &config.NewAPIUser, &config.ClaimQuota,
		&config.KeysAPIURL, &config.KeysAuthorization, &config.GroupID, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		// 返回默认配置
		return &models.AdminConfig{
			NewAPIUser: "1",
			ClaimQuota: 20000000,
			KeysAPIURL: "https://gpt-load.kyx03.de/api/keys/add-async",
			GroupID:    26,
			UpdatedAt:  time.Now().Unix(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateAdminConfig 更新管理员配置
func (db *DB) UpdateAdminConfig(config *models.AdminConfig) error {
	config.UpdatedAt = time.Now().Unix()
	_, err := db.db.Exec(`
		UPDATE admin_config 
		SET session = ?, new_api_user = ?, claim_quota = ?, 
			keys_api_url = ?, keys_authorization = ?, group_id = ?, updated_at = ?
		WHERE id = 1
	`, config.Session, config.NewAPIUser, config.ClaimQuota,
		config.KeysAPIURL, config.KeysAuthorization, config.GroupID, config.UpdatedAt)
	return err
}

// UpdateAdminConfigField 更新管理员配置的单个字段
func (db *DB) UpdateAdminConfigField(field string, value interface{}) error {
	query := fmt.Sprintf(`UPDATE admin_config SET %s = ?, updated_at = ? WHERE id = 1`, field)
	_, err := db.db.Exec(query, value, time.Now().Unix())
	return err
}
