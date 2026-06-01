package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"

	"urlshortener/internal/analytics"
	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/mailer"
	"urlshortener/internal/middleware"
	"urlshortener/internal/modules/ads"
	"urlshortener/internal/modules/auth"
	configmodule "urlshortener/internal/modules/config"
	"urlshortener/internal/modules/links"
	"urlshortener/internal/modules/media"
	oauthmodule "urlshortener/internal/modules/oauth"
	"urlshortener/internal/modules/redirect"
	walletmodule "urlshortener/internal/modules/wallet"
	web3module "urlshortener/internal/modules/web3"
	"urlshortener/internal/repository"
	"urlshortener/internal/storage"
	web3client "urlshortener/internal/web3"
	"urlshortener/pkg/constants"
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

	mailerSvc := mailer.New(cfg.ResendAPIKey, cfg.ResendFrom, cfg.FrontendURL)

	authService := auth.NewAuthService(db, queries, redisClient, cfg, mailerSvc)
	urlService := links.NewURLService(queries, redisClient, cfg)
	analyticsWorker := analytics.NewAnalyticsWorker(queries, 1000)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	go urlService.StartExpiryCleaner(ctx)

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
	oauthService := oauthmodule.NewOAuthService(db, queries, redisClient, cfg)
	oauthHandler := oauthmodule.NewOAuthHandler(oauthService, cfg)
	linksHandler := links.NewLinksHandler(urlService, cfg)
	redirectSvc := redirect.NewRedirectService(urlService, queries, analyticsWorker, geoDB, cfg, db, redisClient.Raw())
	redirectHandler := redirect.NewRedirectHandler(redirectSvc)

	operatorSvc, err := web3client.NewOperatorService(cfg.OperatorPrivateKey, cfg.ContractToken, cfg.NodeRPCURL, big.NewInt(int64(cfg.ChainID)))
	if err != nil {
		zap.L().Fatal("Failed to initialize operator service", zap.Error(err))
	}

	walletSvc := walletmodule.NewWalletService(queries, db, operatorSvc, cfg.PlatformFee, cfg.TokenSymbol)
	walletHandler := walletmodule.NewWalletHandler(walletSvc)

	adSvc := ads.NewAdService(db, queries)
	adHandler := ads.NewAdHandler(adSvc)

	s3Client, err := storage.NewS3Client(cfg)
	if err != nil {
		zap.L().Fatal("Failed to initialize S3 client", zap.Error(err))
	}

	if err := storage.EnsureBucket(ctx, s3Client, cfg); err != nil {
		zap.L().Fatal("Failed to ensure S3 bucket", zap.Error(err))
	}

	mediaSvc := media.NewMediaService(s3Client, cfg.S3Bucket, cfg.S3PublicURL)
	mediaHandler := media.NewMediaHandler(mediaSvc)
	media.StartOrphanCleaner(ctx, db, s3Client, cfg.S3Bucket)

	ethClient := web3client.NewETHClient(cfg, redisClient.Raw())

	faucetAddr := common.HexToAddress(cfg.ContractFaucet)
	faucetSvc, err := web3client.NewFaucetService(cfg.FaucetSignerKey, big.NewInt(int64(cfg.ChainID)), faucetAddr, redisClient.Raw(), cfg.NodeRPCURL)
	if err != nil {
		zap.L().Fatal("Failed to initialize faucet service", zap.Error(err))
	}

	depositHandler := web3module.NewDepositHandlerImpl(db, queries)
	depositListener := web3client.NewListener(ethClient, cfg.ContractPayment, depositHandler, cfg.IsDev())
	web3Svc := web3module.NewWeb3Service(queries, ethClient, faucetSvc, redisClient.Raw())
	web3Handler := web3module.NewWeb3Handler(web3Svc, cfg.IsDev())

	configHandler := configmodule.NewHandler(cfg)

	go depositListener.Start(ctx)
	go faucetSvc.Start(ctx)

	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
		AppName:      "URL Shortener API",
		BodyLimit:    constants.MaxVideoSize,
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
	app.Get("/api/r/:slug/click/:adID", redirectHandler.AdClick)
	app.Get("/api/r/:slug/skip/:adID", redirectHandler.AdSkip)
	app.Post("/api/r/:slug/complete", redirectHandler.AdCompleteFlow)
	app.Get("/api/r/:slug/complete/:adID", redirectHandler.AdComplete)
	app.Get("/api/links/:slug/qr", linksHandler.QRCode)

	api := app.Group("/api", middleware.RateLimiter(redisClient, cfg.RateLimitDefault))

	api.Get("/config", configHandler.GetConfig)
	api.Get("/categories", adHandler.ListCategories)
	api.Get("/ads/types", adHandler.ListAdTypes)

	authGroup := api.Group("/auth", middleware.RateLimiter(redisClient, cfg.RateLimitAuth))
	authGroup.Post("/register", middleware.Turnstile(cfg.TurnstileSecretKey), authHandler.Register)
	authGroup.Post("/login", middleware.Turnstile(cfg.TurnstileSecretKey), authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/send-verification", authHandler.SendVerification)
	authGroup.Post("/verify-email", authHandler.VerifyEmail)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authHandler.ResetPassword)
	authGroup.Get("/google/login", oauthHandler.Login)
	authGroup.Get("/google/callback", oauthHandler.Callback)

	authGroup.Post("/logout", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), authHandler.Logout)
	authGroup.Get("/me", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), authHandler.Me)
	authGroup.Post("/upgrade", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), authHandler.UpgradeToAdvertiser)
	authGroup.Post("/downgrade", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), authHandler.DowngradeToUser)

	walletGroup := api.Group("/wallet", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), middleware.RateLimiter(redisClient, cfg.RateLimitWithdrawal))
	walletGroup.Get("", walletHandler.GetWallet)
	walletGroup.Post("/withdraw", walletHandler.RequestWithdraw)
	walletGroup.Post("/pending", walletHandler.CreatePendingTransaction)
	walletGroup.Get("/ws", websocket.New(walletHandler.ConnectWebSocket))

	faucetGroup := api.Group("/faucet", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries))
	faucetGroup.Post("/claim", middleware.RateLimiter(redisClient, cfg.RateLimitCreate), web3Handler.ClaimFaucet)
	faucetGroup.Post("/confirm", middleware.RateLimiter(redisClient, cfg.RateLimitCreate), web3Handler.ConfirmFaucet)
	faucetGroup.Post("/dev-eth", middleware.RateLimiter(redisClient, cfg.RateLimitCreate), web3Handler.ClaimDevETH)
	faucetGroup.Get("/status", web3Handler.DepositStatus)
	faucetGroup.Get("/history", web3Handler.GetFaucetHistory)

	adsGroup := api.Group("/ads", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), middleware.RequireRole(constants.RoleAdvertiser, constants.RoleAdmin), middleware.RateLimiter(redisClient, cfg.RateLimitMedia))
	adsGroup.Post("/", adHandler.Create)
	adsGroup.Get("/", adHandler.List)
	adsGroup.Get("/:id", adHandler.GetByID)
	adsGroup.Patch("/:id", adHandler.Update)
	adsGroup.Delete("/:id", adHandler.Delete)
	adsGroup.Get("/:id/stats", adHandler.GetStats)
	adsGroup.Post("/:id/topup", adHandler.TopUp)

	mediaGroup := api.Group("/media", middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries), middleware.RequireRole(constants.RoleAdvertiser, constants.RoleAdmin), middleware.RateLimiter(redisClient, cfg.RateLimitMedia))
	mediaGroup.Post("/upload", mediaHandler.Upload)
	mediaGroup.Post("/crop-video", mediaHandler.CropVideo)

	linksGroup := api.Group("/links")
	linksGroup.Use(middleware.JWTAuth(cfg.JWTAccessSecret, redisClient, queries))
	linksGroup.Post("/", middleware.RateLimiter(redisClient, cfg.RateLimitCreate), linksHandler.Create)
	linksGroup.Get("/", linksHandler.List)
	linksGroup.Get("/stats/aggregate", linksHandler.AggregateStats)
	linksGroup.Get("/:slug", linksHandler.Get)
	linksGroup.Get("/:slug/stats", linksHandler.Stats)
	linksGroup.Patch("/:slug", linksHandler.Update)
	linksGroup.Delete("/:slug", linksHandler.Delete)

	// Start WebSocket Wallet Notification Hub
	go walletmodule.GlobalHub.Start(ctx)

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

	depositListener.Stop()
	faucetSvc.Stop()
	faucetSvc.Close()
	operatorSvc.Close()
	stop()

	if err := app.Shutdown(); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}
	zap.L().Info("Server exited")
}
