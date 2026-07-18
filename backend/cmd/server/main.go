package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/config"
	"github.com/ArdyJunata/RitualX/backend/internal/handler"
	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/middleware"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	log := logger.Get()

	log.Info("starting server", "port", cfg.AppPort)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	log.Info("database connected")

	app := fiber.New(fiber.Config{})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(middleware.Trace())

	api := app.Group("/api/v1")
	api.Get("/health", handler.HealthCheck(db))

	// Auth routes
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret)

	auth := api.Group("/auth")
	auth.Post("/register", handler.Register(authService))
	auth.Post("/login", handler.Login(authService))
	auth.Post("/refresh", handler.Refresh(authService))
	auth.Post("/logout", handler.Logout(authService))

	// Routine routes (auth required)
	routineRepo := repository.NewRoutineRepository(db)
	routineService := service.NewRoutineService(routineRepo)

	routines := api.Group("/routines", middleware.RequireAuth(cfg.JWTSecret))
	routines.Post("/", handler.CreateRoutine(routineService))
	routines.Get("/", handler.ListRoutines(routineService))
	routines.Patch("/reorder", handler.ReorderRoutines(routineService))
	routines.Get("/:id", handler.GetRoutine(routineService))
	routines.Put("/:id", handler.UpdateRoutine(routineService))
	routines.Delete("/:id", handler.DeleteRoutine(routineService))

	// Routine log routes
	routineLogRepo := repository.NewRoutineLogRepository(db)
	routineLogService := service.NewRoutineLogService(routineLogRepo, routineRepo)

	routines.Post("/:id/log", handler.LogRoutine(routineLogService))
	routines.Delete("/:id/log/:logId", handler.DeleteRoutineLog(routineLogService))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("shutting down server")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
