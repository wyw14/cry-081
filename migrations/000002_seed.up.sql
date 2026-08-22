INSERT INTO users (id, email, display_name, password_hash, roles, active, created_at, version) VALUES
('demo-author', 'author@example.test', '演示作者', '$2a$10$uTFrQ0VvyJcMvYEtB4wW2Oh2Dg9HrgGC0A2rA9PqLqs5X9WfIgUDm', '["author"]', true, now(), 1),
('demo-editor', 'editor@example.test', '演示编辑', '$2a$10$uTFrQ0VvyJcMvYEtB4wW2Oh2Dg9HrgGC0A2rA9PqLqs5X9WfIgUDm', '["editor"]', true, now(), 1),
('demo-chief', 'chief@example.test', '演示主编', '$2a$10$uTFrQ0VvyJcMvYEtB4wW2Oh2Dg9HrgGC0A2rA9PqLqs5X9WfIgUDm', '["chief_editor"]', true, now(), 1),
('demo-section', 'section@example.test', '栏目管理员', '$2a$10$uTFrQ0VvyJcMvYEtB4wW2Oh2Dg9HrgGC0A2rA9PqLqs5X9WfIgUDm', '["section_manager"]', true, now(), 1)
ON CONFLICT (id) DO NOTHING;
