package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/config"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/publisher"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/relay"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	slog.Info("relay iniciado",
		"brokers", cfg.KafkaBrokers,
		"poll_interval", strconv.Itoa(cfg.PollIntervalMS)+"ms",
	)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("falha ao abrir conexão postgres", "erro", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("falha ao conectar ao postgres", "erro", err)
		os.Exit(1)
	}

	pub := publisher.New(cfg.KafkaBrokers, "pedidos.events")
	defer pub.Close()

	r := relay.New(db, pub, cfg.MaxTentativas)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ticker := time.NewTicker(time.Duration(cfg.PollIntervalMS) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("relay a encerrar")
			return
		case <-ticker.C:
			r.Poll(ctx)
		}
	}
}
