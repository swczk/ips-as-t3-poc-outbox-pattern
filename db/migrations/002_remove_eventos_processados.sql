-- Migration: remove eventos_processados
-- Reason: idempotency moved to Redis (SET NX + TTL)
-- Consumer no longer has PostgreSQL access

DROP TABLE IF EXISTS eventos_processados;
