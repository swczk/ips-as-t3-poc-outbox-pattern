CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS pedidos (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id  UUID          NOT NULL,
    valor       DECIMAL(10,2) NOT NULL,
    estado      TEXT          NOT NULL DEFAULT 'CRIADO',
    criado_em   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tipo          TEXT        NOT NULL,
    payload       JSONB       NOT NULL,
    publicado     BOOLEAN     NOT NULL DEFAULT false,
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    publicado_em  TIMESTAMPTZ,
    tentativas    INT         NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS eventos_processados (
    evento_id     UUID        PRIMARY KEY,
    processado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Relay polls this index; partial keeps it small as rows are marked publicado=true
CREATE INDEX IF NOT EXISTS idx_outbox_nao_publicado
    ON outbox (criado_em ASC)
    WHERE publicado = false;
