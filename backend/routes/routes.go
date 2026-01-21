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

 
func SetupRoutes(router *gin.Engine, db *gorm.DB, cacheManager *cache.CacheManager, logger *zap.Logger) {
	 
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RateLimitMiddleware())

 
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cacheManager, logger)
	userCtrl := controller.NewUserController(userService, logger)

 
	router.GET("/health", userCtrl.Health)

 
	users := router.Group("/users")
	{
		 
		users.POST("", userCtrl.CreateUser)

		 
		users.GET("/:user_id", userCtrl.GetUser)

		 
		users.PUT("/:user_id/rating", userCtrl.UpdateRating)

	 
		users.GET("/:user_id/leaderboard-context", userCtrl.GetLeaderboardAroundUser)

	 
		users.GET("/search", userCtrl.SearchUser)
	}

	 
	leaderboard := router.Group("/leaderboard")
	{
		 
		leaderboard.GET("", userCtrl.GetLeaderboard)
	}
}
