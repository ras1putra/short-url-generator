package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"

	"urlshortener/internal/analytics"
	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/middleware"
	"urlshortener/internal/modules/auth"
	"urlshortener/internal/modules/links"
	"urlshortener/internal/modules/redirect"
	"urlshortener/internal/repository"
	customlogger "urlshortener/pkg/logger"
	"urlshortener/pkg/response"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := customlogger.Init(cfg.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zap.L().Sync()

	zap.L().Info("Configuration loaded successfully", zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	db, err := repository.NewDB(cfg)
	if err != nil {
		zap.L().Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		zap.L().Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Raw().Close()

	queries := repository.New(db)

	authService := auth.NewAuthService(queries, redisClient, cfg)
	urlService := links.NewURLService(queries, redisClient, cfg)
	analyticsWorker := analytics.NewAnalyticsWorker(queries, 1000)

	go urlService.StartExpiryCleaner(context.Background())

	var geoDB *geoip2.Reader
	if cfg.GeoIPDBPath != "" {
		db, err := geoip2.Open(cfg.GeoIPDBPath)
		if err != nil {
			zap.L().Warn("Failed to open GeoIP database, geolocation will be disabled", zap.Error(err))
		} else {
			geoDB = db
			defer geoDB.Close()
		}
	}

	authHandler := auth.NewAuthHandler(authService, cfg)
	linksHandler := links.NewLinksHandler(urlService, cfg)
	redirectHandler := redirect.NewRedirectHandler(urlService, analyticsWorker, geoDB)

	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
		AppName:      "URL Shortener API",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger())

	app.Get("/health", func(c *fiber.Ctx) error {
		return response.OK(c, fiber.Map{"status": "ok", "env": cfg.Env}, "Service is healthy")
	})

	app.Get("/:slug", middleware.RateLimiter(redisClient, cfg.RateLimitRedirect), redirectHandler.Redirect)
	app.Get("/api/links/:slug/qr", linksHandler.QRCode)

	api := app.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)

	protectedAuthGroup := authGroup.Group("/", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient))
	protectedAuthGroup.Post("/logout", authHandler.Logout)

	linksGroup := api.Group("/links")
	linksGroup.Use(middleware.JWTAuth(cfg.JWTAccessSecret, redisClient))
	linksGroup.Post("/", middleware.RateLimiter(redisClient, cfg.RateLimitCreate), linksHandler.Create)
	linksGroup.Get("/", linksHandler.List)
	linksGroup.Get("/stats/aggregate", linksHandler.AggregateStats)
	linksGroup.Get("/:slug", linksHandler.Get)
	linksGroup.Get("/:slug/stats", linksHandler.Stats)
	linksGroup.Patch("/:slug", linksHandler.Update)
	linksGroup.Delete("/:slug", linksHandler.Delete)

	go func() {
		addr := ":" + cfg.Port
		zap.L().Info("Starting server", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			zap.L().Fatal("Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Gracefully shutting down server...")
	if err := app.Shutdown(); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}
	zap.L().Info("Server exited")
}
