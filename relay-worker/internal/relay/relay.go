package relay

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/outbox"
	"github.com/swczk/ips-as-t3-poc-outbox-pattern/relay-worker/internal/publisher"
)

type envelope struct {
	ID      string          `json:"id"`
	Tipo    string          `json:"tipo"`
	Payload json.RawMessage `json:"payload"`
}

type Relay struct {
	db            *sql.DB
	pub           *publisher.Publisher
	maxTentativas int
}

func New(db *sql.DB, pub *publisher.Publisher, maxTentativas int) *Relay {
	return &Relay{db: db, pub: pub, maxTentativas: maxTentativas}
}

func (r *Relay) Poll(ctx context.Context) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("falha ao iniciar transação", "erro", err)
		return
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	eventos, err := outbox.Poll(tx, r.maxTentativas)
	if err != nil {
		slog.Error("falha ao consultar outbox", "erro", err)
		return
	}

	for _, e := range eventos {
		if err := outbox.IncrementTentativas(tx, e.ID); err != nil {
			slog.Error("falha ao incrementar tentativas", "id", e.ID, "erro", err)
			return
		}

		msg, err := json.Marshal(envelope{ID: e.ID, Tipo: e.Tipo, Payload: e.Payload})
		if err != nil {
			slog.Error("falha ao serializar envelope", "id", e.ID, "erro", err)
			continue
		}

		if err := r.pub.Publish(ctx, []byte(e.ID), msg); err != nil {
			slog.Error("falha ao publicar", "id", e.ID, "erro", err)
			continue
		}

		if err := outbox.MarkPublicado(tx, e.ID); err != nil {
			slog.Error("falha ao marcar publicado", "id", e.ID, "erro", err)
			return
		}

		slog.Info("evento publicado", "id", e.ID, "tipo", e.Tipo, "tentativas", e.Tentativas+1)
	}

	if err := tx.Commit(); err != nil {
		slog.Error("falha ao confirmar transação", "erro", err)
		return
	}
	committed = true
}
