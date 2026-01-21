-- Leaderboard System - Database Schema
-- PostgreSQL 12+

-- ========================================
-- Users Table
-- ========================================
CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(255) PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  rating INT NOT NULL DEFAULT 1000
    CHECK (rating >= 100 AND rating <= 5000),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- Indexes for Performance
-- ========================================

-- 1. Index on rating for range queries and ranking
-- Used for: leaderboard queries, rank calculation
-- Selectivity: High (many different ratings)
CREATE INDEX IF NOT EXISTS idx_users_rating 
ON users(rating DESC)
WHERE rating >= 100 AND rating <= 5000;

-- 2. Composite index on (rating DESC, username)
-- Used for: leaderboard pagination with sorting
-- This is a covering index that avoids table access for sorted queries
CREATE INDEX IF NOT EXISTS idx_users_rating_username 
ON users(rating DESC, username);

-- 3. Index on lowercased username for case-insensitive search
-- Used for: search operations with pattern matching
-- Note: GORM uses LOWER() function in WHERE clause
CREATE INDEX IF NOT EXISTS idx_users_username_lower 
ON users(LOWER(username));

-- ========================================
-- Example Queries
-- ========================================

-- Get user by ID (uses primary key index)
-- SELECT * FROM users WHERE id = 'user_123';

-- Get user by username (uses idx_users_username_lower)
-- SELECT * FROM users WHERE LOWER(username) = LOWER('john_doe');

-- Calculate rank for a user (uses idx_users_rating)
-- SELECT COUNT(*) + 1 FROM users WHERE rating > (
--   SELECT rating FROM users WHERE id = 'user_123'
-- );

-- Get leaderboard page (uses idx_users_rating_username)
-- SELECT id, username, rating 
-- FROM users 
-- ORDER BY rating DESC, username ASC 
-- LIMIT 100 OFFSET 0;

-- Search users by username prefix (uses idx_users_username_lower)
-- SELECT * FROM users 
-- WHERE LOWER(username) LIKE LOWER('john%')
-- ORDER BY username ASC
-- LIMIT 10;

-- ========================================
-- Statistics and Maintenance
-- ========================================

-- Analyze table for query planning
-- ANALYZE users;

-- Reindex after heavy updates
-- REINDEX INDEX idx_users_rating;
-- REINDEX INDEX idx_users_username_lower;

-- Check index size
-- SELECT schemaname, tablename, indexname, pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
-- FROM pg_indexes
-- WHERE tablename = 'users';

-- Check table size
-- SELECT pg_size_pretty(pg_total_relation_size('users'));

-- ========================================
-- Constraints
-- ========================================

-- Primary Key: id (unique identifier)
-- Unique Constraint: username (no duplicate usernames)
-- Check Constraint: rating (must be between 100 and 5000)
-- Default Values: created_at, updated_at (CURRENT_TIMESTAMP)
-- Auto Update: updated_at (handled by application layer)

-- ========================================
-- Notes
-- ========================================

-- 1. UUID as primary key (handled by application)
--    Go code: user.ID = uuid.New().String()

-- 2. GORM Auto Migration
--    The table is created automatically by GORM on first run
--    Database.AutoMigrate(&models.User{})

-- 3. Indexes are created by application
--    See: database/database.go -> createIndexes()

-- 4. No foreign keys (leaderboard is self-contained)
--    Users don't reference other tables

-- 5. Timestamp fields
--    created_at: Set once on creation
--    updated_at: Auto-updated on every modification

-- 6. Rating Constraints
--    - Min: 100 (beginner)
--    - Max: 5000 (expert)
--    - Enforced at application level AND database level

-- ========================================
-- Migration Strategy
-- ========================================

-- 1. Create table (automatic via GORM)
-- 2. Create indexes (manual or automatic via application)
-- 3. Verify indexes exist:
--    SELECT * FROM pg_indexes WHERE tablename = 'users';

-- 4. For production:
--    - Backup database before schema changes
--    - Test migrations in staging first
--    - Monitor performance after index creation
--    - Use CONCURRENTLY for non-blocking index creation:
--      CREATE INDEX CONCURRENTLY idx_new ON users(...);

-- ========================================
-- Sample Data for Testing
-- ========================================

-- Insert test users
INSERT INTO users (id, username, rating) VALUES 
  ('user_1', 'alice_player', 5000),
  ('user_2', 'bob_gamer', 4500),
  ('user_3', 'carol_pro', 4500),
  ('user_4', 'david_novice', 2000),
  ('user_5', 'eve_intermediate', 2000)
ON CONFLICT (id) DO NOTHING;

-- Verify rankings (should be: 1, 2, 2, 4, 4)
-- SELECT 
--   id,
--   username,
--   rating,
--   (SELECT COUNT(*) + 1 FROM users u2 WHERE u2.rating > u1.rating) as rank
-- FROM users u1
-- ORDER BY rating DESC, username ASC;
