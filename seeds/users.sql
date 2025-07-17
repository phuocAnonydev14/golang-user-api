-- Seed data for users table
INSERT INTO users (id, username, email, age) VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'john_doe', 'john@example.com', 25),
    ('550e8400-e29b-41d4-a716-446655440001', 'jane_smith', 'jane@example.com', 30),
    ('550e8400-e29b-41d4-a716-446655440002', 'bob_wilson', 'bob@example.com', 22),
    ('550e8400-e29b-41d4-a716-446655440003', 'alice_brown', 'alice@example.com', 28),
    ('550e8400-e29b-41d4-a716-446655440004', 'charlie_davis', 'charlie@example.com', 35)
ON CONFLICT (email) DO NOTHING;