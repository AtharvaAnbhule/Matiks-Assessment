package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"leaderboard-system/service"
)

// UserController handles HTTP requests for user operations
// Implements REST API endpoints
// Responsibilities:
// - Request validation
// - HTTP status codes
// - Error handling
// - Logging
type UserController struct {
	service *service.UserService
	logger  *zap.Logger
}

// NewUserController creates a new controller instance
func NewUserController(service *service.UserService, logger *zap.Logger) *UserController {
	return &UserController{
		service: service,
		logger:  logger,
	}
}

// ErrorResponse represents API error response format
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// SuccessResponse represents successful API response wrapper
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// CreateUser handles POST /users
// Creates a new user with initial rating
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req struct {
		UserID       string `json:"user_id" binding:"required"`
		Username     string `json:"username" binding:"required"`
		InitialRating int32  `json:"initial_rating" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid create user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "INVALID_REQUEST",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	user, err := ctrl.service.CreateUser(c.Request.Context(), req.UserID, req.Username, req.InitialRating)
	if err != nil {
		ctrl.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "CREATE_FAILED",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// GetUser handles GET /users/:user_id
// Returns user info with current rank
func (ctrl *UserController) GetUser(c *gin.Context) {
	userID := c.Param("user_id")

	userDTO, rank, err := ctrl.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		ctrl.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:     "NOT_FOUND",
			Message:   "User not found",
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	// Add rank to response
	response := gin.H{
		"id":       userDTO.ID,
		"username": userDTO.Username,
		"rating":   userDTO.Rating,
		"rank":     rank,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// UpdateRating handles PUT /users/:user_id/rating
// Updates user's rating and triggers rank recalculation
// Non-blocking: returns immediately while cache invalidation happens async
func (ctrl *UserController) UpdateRating(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		Rating int32 `json:"rating" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid rating update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "INVALID_REQUEST",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	userDTO, rank, err := ctrl.service.UpdateUserRating(c.Request.Context(), userID, req.Rating)
	if err != nil {
		ctrl.logger.Error("Failed to update rating", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "UPDATE_FAILED",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	response := gin.H{
		"id":       userDTO.ID,
		"username": userDTO.Username,
		"rating":   userDTO.Rating,
		"rank":     rank,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// SearchUser handles GET /users/search?username=query
// Searches for user by username (case-insensitive)
// Returns user info with rank if found
// Implements debounce on frontend to reduce API calls
func (ctrl *UserController) SearchUser(c *gin.Context) {
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "INVALID_REQUEST",
			Message:   "username parameter required",
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	userDTO, rank, err := ctrl.service.SearchUserByUsername(c.Request.Context(), username)
	if err != nil {
		ctrl.logger.Error("Search failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "SEARCH_FAILED",
			Message:   err.Error(),
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	// Return null user if not found, but 200 OK
	if userDTO == nil {
		c.JSON(http.StatusOK, SuccessResponse{
			Success: true,
			Data: gin.H{
				"user": nil,
				"rank": 0,
				"found": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"user": userDTO,
			"rank": rank,
			"found": true,
		},
	})
}

// GetLeaderboard handles GET /leaderboard?page=1&page_size=100
// Returns paginated leaderboard with ranks
// Pagination params:
// - page: 1-based page number (default: 1)
// - page_size: items per page, max 1000 (default: 100)
func (ctrl *UserController) GetLeaderboard(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "100")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 1000 {
		pageSizeNum = 100
	}

	leaderboard, err := ctrl.service.GetLeaderboard(c.Request.Context(), pageNum, pageSizeNum)
	if err != nil {
		ctrl.logger.Error("Failed to get leaderboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "FETCH_FAILED",
			Message:   "Failed to fetch leaderboard",
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    leaderboard,
	})
}

// GetLeaderboardAroundUser handles GET /users/:user_id/leaderboard-context
// Returns leaderboard entries around user's position
// Shows context: users before and after target user
// Useful for showing how user ranks relative to others
func (ctrl *UserController) GetLeaderboardAroundUser(c *gin.Context) {
	userID := c.Param("user_id")
	contextSize := c.DefaultQuery("context_size", "10")

	contextSizeNum, err := strconv.Atoi(contextSize)
	if err != nil || contextSizeNum < 1 || contextSizeNum > 100 {
		contextSizeNum = 10
	}

	leaderboard, err := ctrl.service.GetLeaderboardAroundUser(c.Request.Context(), userID, contextSizeNum)
	if err != nil {
		ctrl.logger.Error("Failed to get leaderboard context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "FETCH_FAILED",
			Message:   "Failed to fetch leaderboard context",
			Timestamp: time.Now().UTC().String(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    leaderboard,
	})
}

// Health handles GET /health
// Returns service health status
func (ctrl *UserController) Health(c *gin.Context) {
	healthy := ctrl.service.IsHealthy(c.Request.Context())

	if healthy {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().UTC(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"timestamp": time.Now().UTC(),
		})
	}
}
