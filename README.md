# Outbox Pattern — PoC

Proof of concept of the Outbox Pattern with 3 independent services.

```
┌─────────────────┐    ┌──────────────┐    ┌───────────────┐
│  pedidos-service│───▶│  outbox (DB) │───▶│ relay-worker  │───▶ Kafka
│     (Java)      │    │  (Postgres)  │    │    (Go)       │
└─────────────────┘    └──────────────┘    └───────────────┘
                                                                      │
                                                        ┌─────────────▼──────────┐
                                                        │   consumer-service     │
                                                        │     (Node.js/TS)       │
                                                        │  idempotency → Redis   │
                                                        └────────────────────────┘
```

## Services

| Service          | Language                    | Responsibility                                               |
|------------------|-----------------------------|--------------------------------------------------------------|
| pedidos-service  | Java 17 / Spring Boot 3.2.5 | HTTP API + atomic write of order + outbox event              |
| relay-worker     | Go 1.26.3                   | Outbox polling → Kafka publish + circuit breaker + dead letter |
| consumer-service | Node.js >=22.13 / TS        | Consumes Kafka events, idempotency via Redis                 |

## Patterns

| Pattern          | Service(s)                   | Description                                                   |
|------------------|------------------------------|---------------------------------------------------------------|
| Transactional Outbox | pedidos-service          | Order and outbox event written in the same DB transaction     |
| Polling Publisher    | relay-worker             | Polls `outbox` table and publishes pending events to Kafka    |
| Circuit Breaker      | relay-worker             | Lazy ping on publish failure; suspends polling while Kafka is unreachable |
| Dead Letter Queue    | relay-worker             | Events exceeding `MAX_TENTATIVAS` are forwarded to a DLQ topic |
| Idempotent Consumer  | consumer-service         | Redis `SET NX EX` deduplicates redelivered Kafka messages     |

## Quick Start

```bash
cp .env.example .env
docker compose up -d
```

Test:
```bash
curl -s -X POST http://localhost:3000/pedidos \
  -H "Content-Type: application/json" \
  -d '{"clienteId":"11111111-1111-1111-1111-111111111111","valor":25.00}'
```

## Interfaces

| Service       | URL                   |
|---------------|-----------------------|
| Orders API    | http://localhost:3000 |
| Kafka UI      | http://localhost:8080 |
| Redis Insight | http://localhost:5540 |
| Dozzle        | http://localhost:9999 |

## Environment Variables

Copy `.env.example` to `.env`. Key variables:

| Variable               | Service        | Default          |
|------------------------|----------------|------------------|
| `DATABASE_URL`         | pedidos, relay | required         |
| `KAFKA_BROKERS`        | relay, consumer| required         |
| `PEDIDOS_EVENTS_TOPIC` | relay, consumer| required         |
| `DEAD_LETTER_TOPIC`    | relay          | `dead-letter`    |
| `MAX_TENTATIVAS`       | relay          | `5`              |
| `RELAY_BATCH_SIZE`     | relay          | `50`             |
| `POLL_INTERVAL_MS`     | relay          | `1000`           |
| `RECOVERY_INTERVAL_MS` | relay          | `10000`          |
| `CONSUMER_GROUP_ID`    | consumer       | `pedidos-consumer-group` |
| `KAFKA_RETRY_COUNT`    | consumer       | `5`              |
| `REDIS_URL`            | consumer       | `redis://localhost:6379` |
