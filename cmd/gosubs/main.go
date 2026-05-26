package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wallissonmarinho/GoSubs/internal/app"
	appcfg "github.com/wallissonmarinho/GoSubs/internal/app/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := appcfg.Load()
	if err := cfg.Validate(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	engine := app.Wire(cfg)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           engine,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("gosubs listening",
			slog.String("addr", cfg.HTTPAddr()),
			slog.String("cache_dir", cfg.CacheDir),
			slog.String("cache_ttl", cfg.CacheTTL.String()),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
