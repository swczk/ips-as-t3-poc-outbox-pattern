package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/circuit"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/config"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/publisher"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/relay"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	slog.Info("relay worker started",
		"brokers", cfg.KafkaBrokers,
		"pollIntervalMs", cfg.PollIntervalMS,
		"recoveryIntervalMs", cfg.RecoveryIntervalMS,
	)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open postgres connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	pub := publisher.New(cfg.KafkaBrokers, cfg.DeadLetterTopic)
	defer pub.Close()

	cb := circuit.New()
	r := relay.New(db, pub, cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case sig := <-quit:
			slog.Info("relay worker shutting down", "signal", sig.String())
			return
		default:
		}

		if cb.IsOpen() {
			timer := time.NewTimer(time.Duration(cfg.RecoveryIntervalMS) * time.Millisecond)
			select {
			case sig := <-quit:
				timer.Stop()
				slog.Info("relay worker shutting down", "signal", sig.String())
				return
			case <-timer.C:
			}

			if circuit.Ping(cfg.KafkaBrokers) {
				cb.Close()
			} else {
				slog.Warn("kafka still unavailable — waiting for recovery")
			}
			continue
		}

		// ciclo normal — corre em goroutine para permitir graceful shutdown
		done := make(chan bool, 1)
		go func() {
			done <- r.Poll(context.Background())
		}()

		var shouldOpen bool
		select {
		case sig := <-quit:
			// aguarda o ciclo actual antes de sair
			slog.Info("relay worker shutting down — waiting for current cycle", "signal", sig.String())
			<-done
			return
		case shouldOpen = <-done:
		}

		if shouldOpen {
			cb.Open()
			continue
		}

		timer := time.NewTimer(time.Duration(cfg.PollIntervalMS) * time.Millisecond)
		select {
		case sig := <-quit:
			timer.Stop()
			slog.Info("relay worker shutting down", "signal", sig.String())
			return
		case <-timer.C:
		}
	}
}
