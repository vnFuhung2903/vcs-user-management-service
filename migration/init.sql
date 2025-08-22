TRUNCATE TABLE user_scope_mapping RESTART IDENTITY CASCADE;
TRUNCATE TABLE user_scopes RESTART IDENTITY CASCADE;
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

INSERT INTO user_scopes (id, name)
VALUES
(1, 'container:create'),
(2, 'container:view'),
(3, 'container:update'),
(4, 'container:delete'),
(5, 'scope:manage'),
(6, 'user:manage');

INSERT INTO users (id, username, hash, email)
VALUES
('ADMIN', 'admin', '$2a$10$bSo5pXXwb/jcdoZ6RlMdgO9nSNgBKb6DP3MnStijMM2dVHlw.6bl.', 'admin@test.com');

INSERT INTO user_scope_mapping (user_id, user_scope_id)
SELECT 'ADMIN', id FROM user_scopes;