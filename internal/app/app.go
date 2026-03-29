package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mnixry/edgeone-sls-push/internal/config"
	"github.com/mnixry/edgeone-sls-push/internal/edgeone"
	"github.com/mnixry/edgeone-sls-push/internal/httpserver"
	"github.com/mnixry/edgeone-sls-push/internal/sls"
	"github.com/rs/zerolog"
)

func Run(cfg config.CLI, log zerolog.Logger) error {
	auth := edgeone.NewAuthVerifier(edgeone.AuthConfig{
		SecretID:  cfg.EdgeOne.SecretID,
		SecretKey: cfg.EdgeOne.SecretKey,
		MaxSkew:   cfg.EdgeOne.MaxSkew,
	})

	fwd, err := sls.NewForwarder(sls.Config{
		Endpoint:        cfg.SLS.Endpoint,
		AccessKeyID:     cfg.SLS.AccessKeyID,
		AccessKeySecret: cfg.SLS.AccessKeySecret,
		Project:         cfg.SLS.Project,
		LogStore:        cfg.SLS.LogStore,
		Topic:           cfg.SLS.Topic,
		Source:          cfg.SLS.Source,
		LingerMs:        cfg.SLS.LingerMs,
		MaxBatchSize:    cfg.SLS.MaxBatchSize,
		MaxBatchCount:   cfg.SLS.MaxBatchCount,
		Retries:         cfg.SLS.Retries,
		BaseRetryMs:     cfg.SLS.BaseRetryMs,
		MaxRetryMs:      cfg.SLS.MaxRetryMs,
	}, log)
	if err != nil {
		return err
	}
	defer fwd.Close()

	handler := httpserver.NewHandler(auth, fwd, log)

	app := fiber.New(fiber.Config{
		BodyLimit:                 cfg.HTTP.MaxBodyBytes,
		ReadTimeout:               cfg.HTTP.ReadTimeout,
		WriteTimeout:              cfg.HTTP.WriteTimeout,
		EnableIPValidation:        true,
		StreamRequestBody:         false,
		DisableDefaultContentType: true,
	})

	app.Post(cfg.HTTP.Path, handler.Handle)

	app.Get("/healthz", func(c fiber.Ctx) error {
		if err := fwd.Healthy(); err != nil {
			log.Warn().Err(err).Msg("health check failed")
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Str("addr", cfg.HTTP.Addr).Str("path", cfg.HTTP.Path).Msg("HTTP server listening")
		if err := app.Listen(cfg.HTTP.Addr, fiber.ListenConfig{
			DisableStartupMessage: true,
		}); err != nil {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return err
	}

	log.Info().Msg("HTTP server stopped")
	return nil
}
