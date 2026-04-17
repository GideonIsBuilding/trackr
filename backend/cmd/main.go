package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yourname/job-tracker/internal/config"
	"github.com/yourname/job-tracker/internal/db"
	"github.com/yourname/job-tracker/internal/handler"
	authmiddleware "github.com/yourname/job-tracker/internal/middleware"
	prommetrics "github.com/yourname/job-tracker/internal/middleware"
	"github.com/yourname/job-tracker/internal/service"
	"github.com/yourname/job-tracker/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Error("loading config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("database connected")

	// --- Stores ---
	userStore := store.NewUserStore(pool)
	appStore := store.NewApplicationStore(pool)
	contactStore := store.NewContactStore(pool)
	reminderStore := store.NewReminderStore(pool)
	resetStore := store.NewPasswordResetStore(pool)
	analyticsStore := store.NewAnalyticsStore(pool)

	// --- Services ---
	authService := service.NewAuthService(userStore, cfg.JWTSecret, cfg.JWTExpiry)
	emailService := service.NewEmailService(service.EmailConfig{
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUsername: cfg.SMTPUsername,
		SMTPPassword: cfg.SMTPPassword,
		FromAddress:  cfg.FromAddress,
		FromName:     cfg.FromName,
		AppURL:       cfg.AppURL,
	})
	resetService := service.NewPasswordResetService(userStore, resetStore, emailService)
	notifier := service.NewLogNotifier(log)
	reminderEngine := service.NewReminderEngine(reminderStore, notifier, cfg.ReminderCheckInterval, log)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService)
	passwordResetHandler := handler.NewPasswordResetHandler(resetService)
	appHandler := handler.NewApplicationHandler(appStore, contactStore)
	reminderHandler := handler.NewReminderHandler(appStore, reminderStore)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsStore)
	checklistHandler := handler.NewChecklistHandler(appStore)
	extractHandler := handler.NewExtractHandler()

	// --- Router ---
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://yourdomain.com"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Prometheus middleware — records duration and count for every request
	r.Use(prommetrics.PrometheusMiddleware)

	// Metrics endpoint — Prometheus scrapes this every 15s
	r.Handle("/metrics", promhttp.Handler())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/forgot-password", passwordResetHandler.ForgotPassword)
	r.Post("/api/auth/reset-password", passwordResetHandler.ResetPassword)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authmiddleware.Authenticate(cfg.JWTSecret))

		r.Post("/api/applications", appHandler.Create)
		r.Get("/api/applications", appHandler.List)
		r.Get("/api/applications/{id}", appHandler.Get)
		r.Patch("/api/applications/{id}/status", appHandler.UpdateStatus)
		r.Get("/api/applications/{id}/history", appHandler.GetHistory)
		r.Delete("/api/applications/{id}", appHandler.Delete)
		r.Patch("/api/applications/{id}/checklist", checklistHandler.Update)

		r.Put("/api/applications/{id}/reminder", reminderHandler.Configure)
		r.Post("/api/applications/{id}/reminder/snooze", reminderHandler.Snooze)

		r.Get("/api/analytics", analyticsHandler.GetSummary)
		r.Post("/api/extract", extractHandler.Extract)
	})

	// --- Reminder engine ---
	engineCtx, cancelEngine := context.WithCancel(ctx)
	defer cancelEngine()
	go reminderEngine.Run(engineCtx)

	// --- Server ---
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	log.Info("server stopped")
}
