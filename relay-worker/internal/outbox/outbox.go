package outbox

import (
	"database/sql"
	"encoding/json"
)

type Evento struct {
	ID         string
	Tipo       string
	Payload    json.RawMessage
	Tentativas int
}

func Poll(db *sql.DB, maxTentativas, batchSize int) ([]Evento, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.Query(`
		SELECT id, tipo, payload, tentativas
		FROM outbox
		WHERE publicado = false AND tentativas < $1
		ORDER BY criado_em ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, maxTentativas, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventos []Evento
	for rows.Next() {
		var e Evento
		if err := rows.Scan(&e.ID, &e.Tipo, &e.Payload, &e.Tentativas); err != nil {
			return nil, err
		}
		eventos = append(eventos, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return eventos, tx.Commit()
}

func IncrTentativasReturning(db *sql.DB, id string) (int, error) {
	var n int
	err := db.QueryRow(
		`UPDATE outbox SET tentativas = tentativas + 1 WHERE id = $1 RETURNING tentativas`,
		id,
	).Scan(&n)
	return n, err
}

func MarkPublicado(db *sql.DB, id string) error {
	_, err := db.Exec(
		`UPDATE outbox SET publicado = true, publicado_em = NOW(), tentativas = tentativas + 1 WHERE id = $1`,
		id,
	)
	return err
}

func MarkDeadLetter(db *sql.DB, id string) error {
	_, err := db.Exec(
		`UPDATE outbox SET publicado = true, dead_letter = true WHERE id = $1`,
		id,
	)
	return err
}
