-- 插入访问策略，允许 userid = 1 访问 `/api/admin/problem/*` 和 `/api/admin/tags/*`
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', '1', '/api/admin/problem/*', 'CALL', 'allow'),
('p', '1', '/api/admin/tags/*', 'CALL', 'allow');

-- 绑定用户 1 为 admin 角色
INSERT INTO casbin_rule (ptype, v0, v1) VALUES
('g', '1', 'admin');
