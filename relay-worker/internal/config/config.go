package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL        string
	KafkaBrokers       string
	KafkaTopic         string
	DeadLetterTopic    string
	PollIntervalMS     int
	RecoveryIntervalMS int
	MaxTentativas      int
	BatchSize          int
}

func Load() Config {
	return Config{
		DatabaseURL:        mustEnv("DATABASE_URL"),
		KafkaBrokers:       mustEnv("KAFKA_BROKERS"),
		KafkaTopic:         mustEnv("PEDIDOS_EVENTS_TOPIC"),
		DeadLetterTopic:    envStr("DEAD_LETTER_TOPIC", "dead-letter"),
		PollIntervalMS:     envInt("POLL_INTERVAL_MS", 1000),
		RecoveryIntervalMS: envInt("RECOVERY_INTERVAL_MS", 10000),
		MaxTentativas:      envInt("MAX_TENTATIVAS", 5),
		BatchSize:          envInt("RELAY_BATCH_SIZE", 50),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("variável de ambiente obrigatória não definida", "var", key)
		os.Exit(1)
	}
	return v
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
