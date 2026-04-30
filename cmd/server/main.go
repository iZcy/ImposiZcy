package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	kzcydash "github.com/KreaZcy/kzcy-dashboard"
	"github.com/gin-gonic/gin"
	"github.com/iZcy/imposizcy/config"
	"github.com/iZcy/imposizcy/internal/handlers"
	"github.com/iZcy/imposizcy/internal/middleware"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logrus.New()
	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.WithFields(logrus.Fields{
		"port":    cfg.Server.Port,
		"version": "1.0.0",
	}).Info("Starting ImposiZcy")

	mongoService, err := services.NewMongoService(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoService.Close(ctx); err != nil {
			logger.WithError(err).Error("Failed to close MongoDB connection")
		}
	}()

	db := mongoService.GetDatabase()

	templateRepo := repositories.NewTemplateRepository(db)
	renderJobRepo := repositories.NewRenderJobRepository(db)
	imageOutputRepo := repositories.NewImageOutputRepository(db)
	settingsRepo := repositories.NewSettingsRepository(db)
	internalRoleRepo := repositories.NewInternalRoleRepository(db)
	externalRoleRepo := repositories.NewExternalRoleRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	kafkaConnRepo := repositories.NewKafkaConnectionRepository(db)

	rendererService := services.NewRendererService(logger)
	validatorService := services.NewValidatorService(logger)
	imageGenerator := services.NewImageGeneratorService(logger)

	var wsHandler *handlers.WebSocketHandler
	if cfg.Dashboard.Enabled {
		wsHandler = handlers.NewWebSocketHandler(logger)
	}

	templateHandler := handlers.NewTemplateHandler(templateRepo, logger)
	renderHandler := handlers.NewRenderHandler(
		templateRepo, renderJobRepo, imageOutputRepo,
		rendererService, validatorService, imageGenerator,
		wsHandler, logger,
	)
	imageHandler := handlers.NewImageHandler(imageOutputRepo, logger)

	var dashboardHandler *handlers.DashboardHandler
	var rbacHandler *handlers.RBACHandler
	var securityHandler *handlers.SecurityHandler
	var kafkaService *services.KafkaService

	if cfg.Dashboard.Enabled {
		dashboardHandler = handlers.NewDashboardHandler(
			templateRepo, renderJobRepo, imageOutputRepo,
			settingsRepo, kafkaConnRepo,
			rendererService, validatorService, imageGenerator,
			logger,
		)
		rbacHandler = handlers.NewRBACHandler(internalRoleRepo, externalRoleRepo, logger)
		securityHandler = handlers.NewSecurityHandler(apiKeyRepo, settingsRepo, logger)

		kafkaHandler := services.NewKafkaHandler(
			templateRepo, renderJobRepo,
			rendererService, validatorService, imageGenerator,
			logger,
		)

		kafkaService, err = services.NewKafkaService(&cfg.Kafka, logger, kafkaHandler)
		if err != nil {
			logger.WithError(err).Error("Failed to create Kafka service")
		}
		if kafkaService != nil {
			go func() {
				if err := kafkaService.Start(); err != nil {
					logger.WithError(err).Error("Kafka consumer error")
				}
			}()
			logger.Info("Kafka service initialized")
		}
	}

	rateLimiter := middleware.NewRateLimiter(cfg.Security.RateLimitPerMin)
	jwtSecret := cfg.JWT.Secret

	router := gin.New()
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(rateLimiter.Middleware())

	router.GET("/health", gin.WrapH(kzcydash.HealthHandler("ImposiZcy")))

	if securityHandler != nil {
		router.POST("/api/v1/auth/login", securityHandler.Login)
	}

	v1 := router.Group("/api/v1")
	if cfg.Security.APIKeyEnabled {
		v1.Use(middleware.JWTAuthMiddleware(jwtSecret))
	}
	{
		v1.POST("/templates", templateHandler.Create)
		v1.GET("/templates", templateHandler.List)
		v1.GET("/templates/by-slug/:slug", templateHandler.GetBySlug)
		v1.GET("/templates/:id", templateHandler.GetByID)
		v1.PUT("/templates/:id", templateHandler.Update)
		v1.DELETE("/templates/:id", templateHandler.Delete)

		v1.POST("/render", renderHandler.Render)
		v1.POST("/render/async", renderHandler.RenderAsync)
		v1.GET("/render/jobs/:id", renderHandler.GetJobStatus)

		v1.GET("/images/:id", imageHandler.GetByID)
		v1.GET("/images/:id/download", imageHandler.Download)
		v1.GET("/images", imageHandler.List)
		v1.DELETE("/images/:id", imageHandler.Delete)
	}

	if cfg.Dashboard.Enabled {
		dashAuth := middleware.JWTAuthMiddleware(jwtSecret)

		dashboard := router.Group("/dashboard/api")
		dashboard.Use(dashAuth)
		{
			dashboard.GET("/templates", dashboardHandler.ListTemplates)
			dashboard.GET("/templates/:id", dashboardHandler.GetTemplate)
			dashboard.POST("/templates", dashboardHandler.CreateTemplate)
			dashboard.PUT("/templates/:id", dashboardHandler.UpdateTemplate)
			dashboard.DELETE("/templates/:id", dashboardHandler.DeleteTemplate)

			dashboard.GET("/render-jobs", dashboardHandler.ListRenderJobs)
			dashboard.GET("/render-jobs/:id", dashboardHandler.GetRenderJob)

			dashboard.GET("/images", dashboardHandler.ListImages)
			dashboard.DELETE("/images/:id", dashboardHandler.DeleteImage)

			dashboard.GET("/settings", dashboardHandler.GetSettings)
			dashboard.PUT("/settings/:key", dashboardHandler.UpdateSetting)

			dashboard.GET("/kafka/connections", dashboardHandler.ListKafkaConnections)
			dashboard.POST("/kafka/connections", dashboardHandler.CreateKafkaConnection)
			dashboard.PUT("/kafka/connections/:id", dashboardHandler.UpdateKafkaConnection)
			dashboard.DELETE("/kafka/connections/:id", dashboardHandler.DeleteKafkaConnection)
			dashboard.POST("/kafka/connections/:id/start", dashboardHandler.StartKafkaConnection)
			dashboard.POST("/kafka/connections/:id/stop", dashboardHandler.StopKafkaConnection)

			dashboard.GET("/roles/internal", rbacHandler.ListInternalRoles)
			dashboard.POST("/roles/internal", rbacHandler.CreateInternalRole)
			dashboard.PUT("/roles/internal/:id", rbacHandler.UpdateInternalRole)
			dashboard.DELETE("/roles/internal/:id", rbacHandler.DeleteInternalRole)
			dashboard.GET("/roles/external", rbacHandler.ListExternalRoles)
			dashboard.POST("/roles/external", rbacHandler.CreateExternalRole)
			dashboard.DELETE("/roles/external/:id", rbacHandler.DeleteExternalRole)

			dashboard.GET("/security/api-keys", securityHandler.ListAPIKeys)
			dashboard.POST("/security/api-keys", securityHandler.CreateAPIKey)
			dashboard.DELETE("/security/api-keys/:id", securityHandler.DeleteAPIKey)

			dashboard.GET("/ws", wsHandler.HandleConnection)
		}

	}

	isDev := gin.Mode() == gin.DebugMode
	if isDev {
		viteHost := os.Getenv("VITE_HOST")
		if viteHost == "" {
			viteHost = "localhost"
		}
		viteDevServerURL, _ := url.Parse("http://" + viteHost + ":5176")
		viteProxy := httputil.NewSingleHostReverseProxy(viteDevServerURL)
		router.NoRoute(func(c *gin.Context) {
			viteProxy.ServeHTTP(c.Writer, c.Request)
		})
	} else {
		spa := kzcydash.SPAHandler("./public/dist", []string{"/", "/dashboard"})
		router.NoRoute(func(c *gin.Context) {
			spa.ServeHTTP(c.Writer, c.Request)
		})
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if kafkaService != nil {
		if err := kafkaService.Stop(); err != nil {
			logger.WithError(err).Error("Failed to stop Kafka service")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
