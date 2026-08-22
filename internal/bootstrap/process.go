package bootstrap

import (
	"context"
	"errors"
	"net/http"

	"github.com/wyw14/cry-081/internal/config"
	"go.uber.org/zap"
)

// Execute owns process-level resources and returns an operating-system exit code.
func Execute(lifetime context.Context) int {
	cfg, err := config.Load()
	if err != nil {
		return 2
	}
	logger, err := DevelopmentLogger(cfg.Environment)
	if err != nil {
		return 2
	}
	defer func() { _ = logger.Sync() }()

	app, err := Build(lifetime, cfg, logger)
	if err != nil {
		logger.Error("editorial platform bootstrap failed", zap.Error(err))
		return 1
	}
	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           app.Router,
		ReadHeaderTimeout: cfg.HTTP.RequestTimeout,
		ReadTimeout:       cfg.HTTP.RequestTimeout,
		WriteTimeout:      cfg.HTTP.RequestTimeout,
		IdleTimeout:       cfg.HTTP.RequestTimeout * 3,
	}
	serveResult := make(chan error, 1)
	go func() {
		logger.Info("editorial API listening", zap.String("address", cfg.HTTP.Address))
		serveResult <- server.ListenAndServe()
	}()

	failed := false
	select {
	case <-lifetime.Done():
	case serveErr := <-serveResult:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("editorial API stopped unexpectedly", zap.Error(serveErr))
			failed = true
		}
	}
	shutdown, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdown); err != nil {
		logger.Error("HTTP drain did not finish", zap.Error(err))
		failed = true
	}
	if err := app.Close(shutdown); err != nil {
		logger.Error("application resources did not close", zap.Error(err))
		failed = true
	}
	if failed {
		return 1
	}
	return 0
}
