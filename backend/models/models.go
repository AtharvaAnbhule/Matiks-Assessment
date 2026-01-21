package models

import (
	"time"

	
)

 
type User struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Username  string    `gorm:"column:username;uniqueIndex:idx_users_username;type:varchar(255)" json:"username"`
	Rating    int32     `gorm:"column:rating;index:idx_users_rating" json:"rating"` // Range: 100-5000
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

 
func (User) TableName() string {
	return "users"
}


type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int32  `json:"rating"`
	Rank     int64  `json:"rank"` 
}


type LeaderboardEntry struct {
	Rank     int64  `json:"rank"`
	Username string `json:"username"`
	Rating   int32  `json:"rating"`
}


type SearchResult struct {
	User   *UserDTO `json:"user"`
	Rank   int64    `json:"rank"`
	Found  bool     `json:"found"`
	Error  string   `json:"error,omitempty"`
}


type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	HasMore    bool               `json:"has_more"`
}


type RankUpdateEvent struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Rating   int32     `json:"rating"`
	NewRank  int64     `json:"new_rank"`
	Timestamp time.Time `json:"timestamp"`
}
