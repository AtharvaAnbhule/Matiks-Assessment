package models

import (
	"time"

	
)

// User represents a user in the leaderboard system
// Fields are optimized for indexing and fast queries
type User struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Username  string    `gorm:"column:username;uniqueIndex:idx_users_username;type:varchar(255)" json:"username"`
	Rating    int32     `gorm:"column:rating;index:idx_users_rating" json:"rating"` // Range: 100-5000
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// UserDTO represents the data transfer object for user information
// Includes computed rank for API responses
type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int32  `json:"rating"`
	Rank     int64  `json:"rank"` // Computed field
}

// LeaderboardEntry represents a single leaderboard entry
// Used for leaderboard API responses
type LeaderboardEntry struct {
	Rank     int64  `json:"rank"`
	Username string `json:"username"`
	Rating   int32  `json:"rating"`
}

// SearchResult represents search results with pagination
type SearchResult struct {
	User   *UserDTO `json:"user"`
	Rank   int64    `json:"rank"`
	Found  bool     `json:"found"`
	Error  string   `json:"error,omitempty"`
}

// LeaderboardResponse represents paginated leaderboard data
type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	HasMore    bool               `json:"has_more"`
}

// RankUpdateEvent represents a rank update for real-time updates
// Can be used for WebSocket broadcasts
type RankUpdateEvent struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Rating   int32     `json:"rating"`
	NewRank  int64     `json:"new_rank"`
	Timestamp time.Time `json:"timestamp"`
}
