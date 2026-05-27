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

func Poll(tx *sql.Tx, maxTentativas, batchSize int) ([]Evento, error) {
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
	return eventos, rows.Err()
}

func IncrementTentativas(tx *sql.Tx, id string) error {
	_, err := tx.Exec(`UPDATE outbox SET tentativas = tentativas + 1 WHERE id = $1`, id)
	return err
}

func MarkPublicado(tx *sql.Tx, id string) error {
	_, err := tx.Exec(`UPDATE outbox SET publicado = true, publicado_em = NOW() WHERE id = $1`, id)
	return err
}
