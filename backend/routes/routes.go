package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"leaderboard-system/cache"
	"leaderboard-system/controller"
	"leaderboard-system/middleware"
	"leaderboard-system/repository"
	"leaderboard-system/service"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes and middleware
// Dependency injection pattern for clean architecture
func SetupRoutes(router *gin.Engine, db *gorm.DB, cacheManager *cache.CacheManager, logger *zap.Logger) {
	// Middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RateLimitMiddleware())

	// Initialize repository, service, and controller
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cacheManager, logger)
	userCtrl := controller.NewUserController(userService, logger)

	// Health check
	router.GET("/health", userCtrl.Health)

	// User endpoints
	users := router.Group("/users")
	{
		// Create user
		users.POST("", userCtrl.CreateUser)

		// Get specific user with rank
		users.GET("/:user_id", userCtrl.GetUser)

		// Update user rating (triggers rank recalculation)
		users.PUT("/:user_id/rating", userCtrl.UpdateRating)

		// Get leaderboard context around user
		users.GET("/:user_id/leaderboard-context", userCtrl.GetLeaderboardAroundUser)

		// Search user by username
		users.GET("/search", userCtrl.SearchUser)
	}

	// Leaderboard endpoints
	leaderboard := router.Group("/leaderboard")
	{
		// Get paginated leaderboard
		leaderboard.GET("", userCtrl.GetLeaderboard)
	}
}
