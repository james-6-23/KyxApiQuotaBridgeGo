-- ========================================
-- KyxApiQuotaBridge 数据库初始化脚本
-- ========================================
-- 数据库版本: PostgreSQL 15+
-- 创建时间: 2024-01-01
-- 说明: 初始化所有表结构、索引、约束和默认数据
-- ========================================

-- 设置客户端编码
SET client_encoding = 'UTF8';

-- 创建扩展（如果需要）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- 1. 用户表 (users)
-- ========================================
-- 存储已绑定的用户信息
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    linux_do_id VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    kyx_user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_linux_do_id ON users(linux_do_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_kyx_user_id ON users(kyx_user_id);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 添加注释
COMMENT ON TABLE users IS '用户表，存储已绑定的Linux Do和公益站账号信息';
COMMENT ON COLUMN users.id IS '主键ID';
COMMENT ON COLUMN users.linux_do_id IS 'Linux Do 用户ID，唯一标识';
COMMENT ON COLUMN users.username IS '公益站用户名';
COMMENT ON COLUMN users.kyx_user_id IS '公益站用户ID';
COMMENT ON COLUMN users.created_at IS '创建时间（首次绑定时间）';
COMMENT ON COLUMN users.updated_at IS '最后更新时间';

-- ========================================
-- 2. 领取记录表 (claim_records)
-- ========================================
-- 存储用户每日领取额度的记录
CREATE TABLE IF NOT EXISTS claim_records (
    id SERIAL PRIMARY KEY,
    linux_do_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    quota_added BIGINT NOT NULL CHECK (quota_added > 0),
    claim_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一约束：每个用户每天只能领取一次
CREATE UNIQUE INDEX IF NOT EXISTS idx_claim_records_unique ON claim_records(linux_do_id, claim_date);

-- 创建其他索引
CREATE INDEX IF NOT EXISTS idx_claim_records_linux_do_id ON claim_records(linux_do_id);
CREATE INDEX IF NOT EXISTS idx_claim_records_claim_date ON claim_records(claim_date);
CREATE INDEX IF NOT EXISTS idx_claim_records_created_at ON claim_records(created_at);
CREATE INDEX IF NOT EXISTS idx_claim_records_username ON claim_records(username);

-- 添加注释
COMMENT ON TABLE claim_records IS '领取记录表，记录用户每日领取额度的历史';
COMMENT ON COLUMN claim_records.id IS '主键ID';
COMMENT ON COLUMN claim_records.linux_do_id IS 'Linux Do 用户ID';
COMMENT ON COLUMN claim_records.username IS '用户名';
COMMENT ON COLUMN claim_records.quota_added IS '添加的额度数值';
COMMENT ON COLUMN claim_records.claim_date IS '领取日期';
COMMENT ON COLUMN claim_records.created_at IS '记录创建时间';

-- ========================================
-- 3. 投喂记录表 (donate_records)
-- ========================================
-- 存储用户投喂ModelScope Key的记录
CREATE TABLE IF NOT EXISTS donate_records (
    id SERIAL PRIMARY KEY,
    linux_do_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    keys_count INTEGER NOT NULL CHECK (keys_count > 0),
    total_quota_added BIGINT NOT NULL CHECK (total_quota_added >= 0),
    push_status VARCHAR(20) DEFAULT 'success' CHECK (push_status IN ('success', 'failed')),
    push_message TEXT,
    failed_keys JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_donate_records_linux_do_id ON donate_records(linux_do_id);
CREATE INDEX IF NOT EXISTS idx_donate_records_created_at ON donate_records(created_at);
CREATE INDEX IF NOT EXISTS idx_donate_records_username ON donate_records(username);
CREATE INDEX IF NOT EXISTS idx_donate_records_push_status ON donate_records(push_status);
CREATE INDEX IF NOT EXISTS idx_donate_records_date ON donate_records(DATE(created_at));

-- 创建GIN索引用于JSONB字段
CREATE INDEX IF NOT EXISTS idx_donate_records_failed_keys ON donate_records USING GIN (failed_keys);

-- 添加注释
COMMENT ON TABLE donate_records IS '投喂记录表，记录用户投喂ModelScope Key的历史';
COMMENT ON COLUMN donate_records.id IS '主键ID';
COMMENT ON COLUMN donate_records.linux_do_id IS 'Linux Do 用户ID';
COMMENT ON COLUMN donate_records.username IS '用户名';
COMMENT ON COLUMN donate_records.keys_count IS '投喂的Key数量';
COMMENT ON COLUMN donate_records.total_quota_added IS '总共添加的额度';
COMMENT ON COLUMN donate_records.push_status IS '推送状态：success/failed';
COMMENT ON COLUMN donate_records.push_message IS '推送消息';
COMMENT ON COLUMN donate_records.failed_keys IS '推送失败的Keys（JSON数组）';
COMMENT ON COLUMN donate_records.created_at IS '记录创建时间';

-- ========================================
-- 4. 已使用Keys表 (used_keys)
-- ========================================
-- 存储已经使用过的ModelScope Keys（防止重复使用）
CREATE TABLE IF NOT EXISTS used_keys (
    key_hash VARCHAR(64) PRIMARY KEY,
    full_key TEXT NOT NULL,
    linux_do_id VARCHAR(100),
    username VARCHAR(100),
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_used_keys_linux_do_id ON used_keys(linux_do_id);
CREATE INDEX IF NOT EXISTS idx_used_keys_used_at ON used_keys(used_at);
CREATE INDEX IF NOT EXISTS idx_used_keys_username ON used_keys(username);

-- 添加注释
COMMENT ON TABLE used_keys IS '已使用Keys表，防止Key被重复提交';
COMMENT ON COLUMN used_keys.key_hash IS 'Key的SHA256哈希值，作为主键';
COMMENT ON COLUMN used_keys.full_key IS '完整的Key（加密存储）';
COMMENT ON COLUMN used_keys.linux_do_id IS '使用该Key的用户Linux Do ID';
COMMENT ON COLUMN used_keys.username IS '使用该Key的用户名';
COMMENT ON COLUMN used_keys.used_at IS 'Key使用时间';

-- ========================================
-- 5. 管理员配置表 (admin_config)
-- ========================================
-- 存储系统管理员配置信息（单行表）
CREATE TABLE IF NOT EXISTS admin_config (
    id INTEGER PRIMARY KEY DEFAULT 1,
    session TEXT,
    new_api_user VARCHAR(100) DEFAULT '1',
    claim_quota BIGINT DEFAULT 20000000 CHECK (claim_quota > 0),
    keys_api_url TEXT,
    keys_authorization TEXT,
    group_id INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row_only CHECK (id = 1)
);

-- 添加注释
COMMENT ON TABLE admin_config IS '管理员配置表（单行表）';
COMMENT ON COLUMN admin_config.id IS '主键ID（固定为1）';
COMMENT ON COLUMN admin_config.session IS '公益站API Session';
COMMENT ON COLUMN admin_config.new_api_user IS 'new-api-user 请求头';
COMMENT ON COLUMN admin_config.claim_quota IS '每日领取额度';
COMMENT ON COLUMN admin_config.keys_api_url IS 'Keys推送API地址';
COMMENT ON COLUMN admin_config.keys_authorization IS 'Keys推送授权令牌';
COMMENT ON COLUMN admin_config.group_id IS 'Keys推送的分组ID';
COMMENT ON COLUMN admin_config.updated_at IS '最后更新时间';

-- 插入默认配置
INSERT INTO admin_config (id, claim_quota, new_api_user)
VALUES (1, 20000000, '1')
ON CONFLICT (id) DO NOTHING;

-- ========================================
-- 6. 会话表 (sessions)
-- ========================================
-- 存储用户和管理员的会话信息
CREATE TABLE IF NOT EXISTS sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    data JSONB NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);

-- 创建GIN索引用于JSONB字段
CREATE INDEX IF NOT EXISTS idx_sessions_data ON sessions USING GIN (data);

-- 添加注释
COMMENT ON TABLE sessions IS '会话表，存储用户和管理员的登录会话';
COMMENT ON COLUMN sessions.session_id IS '会话ID（UUID或随机字符串）';
COMMENT ON COLUMN sessions.data IS '会话数据（JSON格式）';
COMMENT ON COLUMN sessions.expires_at IS '过期时间';
COMMENT ON COLUMN sessions.created_at IS '创建时间';

-- ========================================
-- 7. 创建更新时间自动触发器
-- ========================================

-- 创建更新时间戳函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为 users 表添加触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为 admin_config 表添加触发器
DROP TRIGGER IF EXISTS update_admin_config_updated_at ON admin_config;
CREATE TRIGGER update_admin_config_updated_at
    BEFORE UPDATE ON admin_config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 8. 创建清理过期会话的函数
-- ========================================

CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_sessions() IS '清理过期的会话记录';

-- ========================================
-- 9. 创建统计视图
-- ========================================

-- 用户统计视图
CREATE OR REPLACE VIEW user_statistics AS
SELECT
    u.linux_do_id,
    u.username,
    u.created_at as register_time,
    COUNT(DISTINCT cr.id) as total_claims,
    COALESCE(SUM(cr.quota_added), 0) as total_claim_quota,
    COUNT(DISTINCT dr.id) as total_donates,
    COALESCE(SUM(dr.keys_count), 0) as total_keys_donated,
    COALESCE(SUM(dr.total_quota_added), 0) as total_donate_quota,
    COALESCE(SUM(cr.quota_added), 0) + COALESCE(SUM(dr.total_quota_added), 0) as total_quota
FROM users u
LEFT JOIN claim_records cr ON u.linux_do_id = cr.linux_do_id
LEFT JOIN donate_records dr ON u.linux_do_id = dr.linux_do_id
GROUP BY u.linux_do_id, u.username, u.created_at;

COMMENT ON VIEW user_statistics IS '用户统计视图，包含领取和投喂的汇总数据';

-- 每日统计视图
CREATE OR REPLACE VIEW daily_statistics AS
SELECT
    claim_date as date,
    COUNT(DISTINCT linux_do_id) as unique_users,
    COUNT(*) as total_claims,
    SUM(quota_added) as total_quota
FROM claim_records
GROUP BY claim_date
ORDER BY claim_date DESC;

COMMENT ON VIEW daily_statistics IS '每日统计视图，按日期汇总领取数据';

-- 投喂统计视图
CREATE OR REPLACE VIEW donate_statistics AS
SELECT
    DATE(created_at) as date,
    COUNT(DISTINCT linux_do_id) as unique_users,
    COUNT(*) as total_donates,
    SUM(keys_count) as total_keys,
    SUM(total_quota_added) as total_quota,
    COUNT(CASE WHEN push_status = 'success' THEN 1 END) as success_count,
    COUNT(CASE WHEN push_status = 'failed' THEN 1 END) as failed_count
FROM donate_records
GROUP BY DATE(created_at)
ORDER BY DATE(created_at) DESC;

COMMENT ON VIEW donate_statistics IS '投喂统计视图，按日期汇总投喂数据';

-- ========================================
-- 10. 创建性能优化函数
-- ========================================

-- 更新表统计信息
CREATE OR REPLACE FUNCTION update_table_statistics()
RETURNS VOID AS $$
BEGIN
    ANALYZE users;
    ANALYZE claim_records;
    ANALYZE donate_records;
    ANALYZE used_keys;
    ANALYZE sessions;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_table_statistics() IS '更新所有表的统计信息，优化查询性能';

-- ========================================
-- 11. 数据库维护任务（定时清理）
-- ========================================

-- 注意：以下是建议的维护任务，需要配置 pg_cron 或外部定时任务

-- 清理30天前的过期会话（可以通过应用层或定时任务执行）
-- SELECT cleanup_expired_sessions();

-- 清理90天前的领取记录（可选，如果需要保留历史数据则不执行）
-- DELETE FROM claim_records WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '90 days';

-- 更新统计信息（建议每天执行一次）
-- SELECT update_table_statistics();

-- ========================================
-- 12. 权限设置（可选）
-- ========================================

-- 如果需要创建只读用户
-- CREATE USER readonly_user WITH PASSWORD 'your_password';
-- GRANT CONNECT ON DATABASE kyxquota TO readonly_user;
-- GRANT USAGE ON SCHEMA public TO readonly_user;
-- GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;
-- GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO readonly_user;

-- ========================================
-- 13. 数据验证和完整性检查
-- ========================================

-- 创建数据验证函数
CREATE OR REPLACE FUNCTION validate_data_integrity()
RETURNS TABLE(
    check_name TEXT,
    status TEXT,
    details TEXT
) AS $$
BEGIN
    -- 检查孤立的领取记录
    RETURN QUERY
    SELECT
        '孤立的领取记录'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        '找到 ' || COUNT(*) || ' 条孤立记录'::TEXT
    FROM claim_records cr
    LEFT JOIN users u ON cr.linux_do_id = u.linux_do_id
    WHERE u.linux_do_id IS NULL;

    -- 检查孤立的投喂记录
    RETURN QUERY
    SELECT
        '孤立的投喂记录'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        '找到 ' || COUNT(*) || ' 条孤立记录'::TEXT
    FROM donate_records dr
    LEFT JOIN users u ON dr.linux_do_id = u.linux_do_id
    WHERE u.linux_do_id IS NULL;

    -- 检查过期会话数量
    RETURN QUERY
    SELECT
        '过期会话'::TEXT,
        'INFO'::TEXT,
        '找到 ' || COUNT(*) || ' 个过期会话'::TEXT
    FROM sessions
    WHERE expires_at < CURRENT_TIMESTAMP;

    -- 检查表记录数
    RETURN QUERY
    SELECT
        '总用户数'::TEXT,
        'INFO'::TEXT,
        COUNT(*)::TEXT || ' 个用户'
    FROM users;

    RETURN QUERY
    SELECT
        '总领取记录'::TEXT,
        'INFO'::TEXT,
        COUNT(*)::TEXT || ' 条记录'
    FROM claim_records;

    RETURN QUERY
    SELECT
        '总投喂记录'::TEXT,
        'INFO'::TEXT,
        COUNT(*)::TEXT || ' 条记录'
    FROM donate_records;

    RETURN QUERY
    SELECT
        '已使用Keys'::TEXT,
        'INFO'::TEXT,
        COUNT(*)::TEXT || ' 个Keys'
    FROM used_keys;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_data_integrity() IS '验证数据完整性，检查孤立记录和统计信息';

-- ========================================
-- 14. 初始化完成信息
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'KyxApiQuotaBridge 数据库初始化完成！';
    RAISE NOTICE '========================================';
    RAISE NOTICE '已创建的表：';
    RAISE NOTICE '  1. users - 用户表';
    RAISE NOTICE '  2. claim_records - 领取记录表';
    RAISE NOTICE '  3. donate_records - 投喂记录表';
    RAISE NOTICE '  4. used_keys - 已使用Keys表';
    RAISE NOTICE '  5. admin_config - 管理员配置表';
    RAISE NOTICE '  6. sessions - 会话表';
    RAISE NOTICE '';
    RAISE NOTICE '已创建的视图：';
    RAISE NOTICE '  1. user_statistics - 用户统计视图';
    RAISE NOTICE '  2. daily_statistics - 每日统计视图';
    RAISE NOTICE '  3. donate_statistics - 投喂统计视图';
    RAISE NOTICE '';
    RAISE NOTICE '已创建的函数：';
    RAISE NOTICE '  1. cleanup_expired_sessions() - 清理过期会话';
    RAISE NOTICE '  2. update_table_statistics() - 更新统计信息';
    RAISE NOTICE '  3. validate_data_integrity() - 验证数据完整性';
    RAISE NOTICE '';
    RAISE NOTICE '建议定期执行：';
    RAISE NOTICE '  - SELECT cleanup_expired_sessions();';
    RAISE NOTICE '  - SELECT update_table_statistics();';
    RAISE NOTICE '  - SELECT * FROM validate_data_integrity();';
    RAISE NOTICE '========================================';
END $$;

-- 执行一次数据完整性检查
SELECT * FROM validate_data_integrity();
