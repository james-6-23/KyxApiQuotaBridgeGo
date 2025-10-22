-- ========================================
-- 修复 admin_config 表中的 NULL 字段
-- ========================================
-- 说明: 将 NULL 值更新为空字符串，以兼容新的字段定义
-- ========================================

-- 更新 session 字段的 NULL 值为空字符串
UPDATE admin_config 
SET session = COALESCE(session, '')
WHERE session IS NULL;

-- 更新 new_api_user 字段的 NULL 值为 '1'
UPDATE admin_config 
SET new_api_user = COALESCE(new_api_user, '1')
WHERE new_api_user IS NULL;

-- 更新 keys_api_url 字段的 NULL 值为空字符串
UPDATE admin_config 
SET keys_api_url = COALESCE(keys_api_url, '')
WHERE keys_api_url IS NULL;

-- 更新 keys_authorization 字段的 NULL 值为空字符串
UPDATE admin_config 
SET keys_authorization = COALESCE(keys_authorization, '')
WHERE keys_authorization IS NULL;

-- 验证更新结果
SELECT 
    id,
    CASE WHEN session IS NULL THEN 'NULL' ELSE 'NOT NULL' END as session_status,
    CASE WHEN new_api_user IS NULL THEN 'NULL' ELSE 'NOT NULL' END as new_api_user_status,
    CASE WHEN keys_api_url IS NULL THEN 'NULL' ELSE 'NOT NULL' END as keys_api_url_status,
    CASE WHEN keys_authorization IS NULL THEN 'NULL' ELSE 'NOT NULL' END as keys_authorization_status
FROM admin_config;

-- 提示信息
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'admin_config 表中的 NULL 字段已修复';
    RAISE NOTICE '========================================';
END $$;