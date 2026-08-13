-- ============================================================
-- 迁移脚本：将联系人上的"关系类标签"迁移为真实的 relations 记录
-- 背景：改造后，亲密度(strength)与关系标签(tags)都属于"关系"，
--       不再挂在单个联系人身上。
-- 一次性脚本，可安全重复执行（INSERT IGNORE / ON DUPLICATE KEY UPDATE）。
-- ============================================================

-- 1) 为 contacts.referrer_id 引荐链自动生成"引荐"关系（亲密度默认 5）
--    from = 引荐人(引荐链上级)，to = 被引荐联系人
INSERT INTO relations (id, tenant_id, from_person_id, to_person_id, type, tags, strength, note, created_by, created_at, updated_at)
SELECT
    UUID(),
    c.tenant_id,
    c.referrer_id,
    c.person_id,
    'referral',
    '["引荐"]',
    5,
    '由引荐链自动迁移',
    c.created_by,
    c.created_at,
    c.updated_at
FROM contacts c
WHERE c.referrer_id IS NOT NULL AND c.referrer_id != ''
  AND NOT EXISTS (
    SELECT 1 FROM relations r
    WHERE r.from_person_id = c.referrer_id
      AND r.to_person_id = c.person_id
      AND r.type = 'referral'
  );

-- 2) 把联系人身上的关系类标签迁移为关系记录
--    from = 该租户第一个创建的用户（"我"），to = 联系人
--    关系类型取自定义(custom)，tags 保留原始联系人标签
INSERT INTO relations (id, tenant_id, from_person_id, to_person_id, type, tags, strength, note, created_by, created_at, updated_at)
SELECT
    UUID(),
    c.tenant_id,
    (SELECT u.id FROM users u WHERE u.tenant_id = c.tenant_id ORDER BY u.created_at LIMIT 1),
    c.person_id,
    'custom',
    c.tags,
    5,
    '由联系人标签自动迁移',
    c.created_by,
    c.created_at,
    c.updated_at
FROM contacts c
WHERE c.tags IS NOT NULL AND c.tags != '[]' AND c.tags != ''
  AND NOT EXISTS (
    SELECT 1 FROM relations r
    WHERE r.from_person_id = (SELECT u.id FROM users u WHERE u.tenant_id = c.tenant_id ORDER BY u.created_at LIMIT 1)
      AND r.to_person_id = c.person_id
  );

-- 3) 校验
-- SELECT type, strength, COUNT(*) FROM relations GROUP BY type, strength ORDER BY strength;
