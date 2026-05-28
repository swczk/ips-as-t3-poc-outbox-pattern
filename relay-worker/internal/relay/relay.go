package relay

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/circuit"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/config"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/outbox"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/publisher"
)

type Relay struct {
	db              *sql.DB
	pub             *publisher.Publisher
	brokers         string
	pedidosTopic    string
	deadLetterTopic string
	maxTentativas   int
	batchSize       int
}

func New(db *sql.DB, pub *publisher.Publisher, cfg config.Config) *Relay {
	return &Relay{
		db:              db,
		pub:             pub,
		brokers:         cfg.KafkaBrokers,
		pedidosTopic:    cfg.KafkaTopic,
		deadLetterTopic: cfg.DeadLetterTopic,
		maxTentativas:   cfg.MaxTentativas,
		batchSize:       cfg.BatchSize,
	}
}

func (r *Relay) Poll(ctx context.Context) bool {
	eventos, err := outbox.Poll(r.db, r.maxTentativas, r.batchSize)
	if err != nil {
		slog.Error("failed to query outbox", "error", err)
		return false
	}

	if len(eventos) == 0 {
		slog.Info("no pending events")
		return false
	}

	for _, e := range eventos {
		if shouldOpen := r.processarEvento(ctx, e); shouldOpen {
			return true
		}
	}
	return false
}

func (r *Relay) processarEvento(ctx context.Context, e outbox.Evento) bool {
	topic := r.topicPara(e.Tipo)

	err := r.pub.Publish(ctx, topic, []byte(e.ID), e.Payload)
	if err == nil {
		if dbErr := outbox.MarkPublicado(r.db, e.ID); dbErr != nil {
			slog.Error("failed to mark event as published", "eventId", e.ID, "error", dbErr)
		} else {
			slog.Info("event published",
				"eventId", e.ID, "tipo", e.Tipo, "tentativas", e.Tentativas+1, "topic", topic)
		}
		return false
	}

	motivo := err.Error()

	if e.Tentativas == 0 {
		novas, dbErr := outbox.IncrTentativasReturning(r.db, e.ID)
		if dbErr != nil {
			slog.Error("failed to increment tentativas", "eventId", e.ID, "error", dbErr)
			novas = 1
		}

		if !circuit.Ping(r.brokers) {
			slog.Error("failed to publish event",
				"eventId", e.ID, "tipo", e.Tipo, "tentativas", novas, "error", motivo)
			return true
		}

		slog.Error("failed to publish event",
			"eventId", e.ID, "tipo", e.Tipo, "tentativas", novas, "error", motivo)

		if novas >= r.maxTentativas {
			return r.enviarParaDeadLetter(ctx, e, novas, motivo)
		}
		return false
	}

	if !circuit.Ping(r.brokers) {
		slog.Error("failed to publish event",
			"eventId", e.ID, "tipo", e.Tipo, "tentativas", e.Tentativas, "error", motivo)
		return true
	}

	novas, dbErr := outbox.IncrTentativasReturning(r.db, e.ID)
	if dbErr != nil {
		slog.Error("failed to increment tentativas", "eventId", e.ID, "error", dbErr)
		novas = e.Tentativas + 1
	}

	slog.Error("failed to publish event",
		"eventId", e.ID, "tipo", e.Tipo, "tentativas", novas, "error", motivo)

	if novas >= r.maxTentativas {
		return r.enviarParaDeadLetter(ctx, e, novas, motivo)
	}
	return false
}

func (r *Relay) enviarParaDeadLetter(ctx context.Context, e outbox.Evento, tentativas int, motivo string) bool {
	dlPayload, _ := json.Marshal(map[string]any{
		"eventoOriginal": e.Payload,
		"tipo":           e.Tipo,
		"tentativas":     tentativas,
		"motivo":         motivo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})

	dlErr := r.pub.PublishDeadLetter(ctx, []byte(e.ID), dlPayload)
	if dlErr != nil {
		dlMotivo := dlErr.Error()
		if !circuit.Ping(r.brokers) {
			slog.Error("failed to send event to dead-letter",
				"eventId", e.ID, "error", dlMotivo)
			return true
		}
		slog.Error("failed to send event to dead-letter",
			"eventId", e.ID, "error", dlMotivo)
	}

	if dbErr := outbox.MarkDeadLetter(r.db, e.ID); dbErr != nil {
		slog.Error("failed to mark event as dead-letter", "eventId", e.ID, "error", dbErr)
	} else {
		slog.Warn("event sent to dead-letter after max retries",
			"eventId", e.ID, "tipo", e.Tipo, "tentativas", tentativas, "deadLetterTopic", r.deadLetterTopic)
	}
	return false
}

func (r *Relay) topicPara(tipo string) string {
	if tipo == "PedidoCriado" {
		return r.pedidosTopic
	}
	return strings.ToLower(tipo)
}
