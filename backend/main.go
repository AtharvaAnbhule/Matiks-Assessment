package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"leaderboard-system/cache"
	"leaderboard-system/config"
	"leaderboard-system/database"
	"leaderboard-system/routes"
	"gorm.io/gorm/logger"
)

func main() {
	 
	_ = godotenv.Load()

 
	cfg := config.GetConfig()

	 
	var log *zap.Logger
	var err error
	if cfg.Server.Env == "production" {
		log, err = zap.NewProduction()
	} else {
		log, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	log.Info("Starting leaderboard service",
		zap.String("environment", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	 
	logLevel := logger.Silent
	if cfg.Server.Env != "production" {
		logLevel = logger.Info
	}

	db, err := database.InitDB(&cfg.Database, logLevel)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}

	log.Info("Database connected")

	 
	if cfg.Server.Env != "production" {
		if err := database.SeedData(db); err != nil {
			log.Warn("Failed to seed data", zap.Error(err))
		} else {
			log.Info("Database seeded with initial data")
		}
	}

 
	cacheManager, err := cache.NewCacheManager(&cfg.Redis)
	if err != nil {
		log.Fatal("Failed to initialize cache", zap.Error(err))
	}
	defer cacheManager.Close()

	log.Info("Cache connected")

 
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

 
	routes.SetupRoutes(router, db, cacheManager, log)

 
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

 
	go func() {
		log.Info("Server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", zap.Error(err))
		}
	}()

 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

 
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped")
}
