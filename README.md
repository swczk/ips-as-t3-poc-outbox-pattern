# Outbox Pattern — PoC

Prova de conceito do Outbox Pattern com 3 serviços independentes.

## Serviços

| Serviço           | Linguagem  | Responsabilidade                        |
|-------------------|------------|-----------------------------------------|
| pedidos-service   | Java       | API HTTP + persistência atómica         |
| relay-worker      | Go         | Lê outbox e publica no Kafka            |
| consumer-service  | Node.js    | Consome eventos, garante idempotência   |

## Arranque rápido

### Pré-requisitos
- Docker + Docker Compose
- Java 17+, Maven 3.9+
- Go 1.21+
- Node.js 22+, pnpm

### Infra
```bash
docker compose up -d
```

### Schema
```bash
PGPASSWORD=poc_pass psql -h localhost -p 5432 -U poc_user -d poc_db \
  -f db/schema.sql
```

### Serviços (3 terminais)
```bash
# Terminal 1 — Pedidos (Java)
cd pedidos-service
DATABASE_URL=jdbc:postgresql://localhost:5432/poc_db?user=poc_user&password=poc_pass \
  mvn spring-boot:run

# Terminal 2 — Relay (Go)
cd relay-worker && go run main.go

# Terminal 3 — Consumer (Node.js)
cd consumer-service && pnpm install && pnpm run dev
```

### Testar
```bash
curl -s -X POST http://localhost:3000/pedidos \
  -H "Content-Type: application/json" \
  -d '{"clienteId":"11111111-1111-1111-1111-111111111111","valor":25.00}'
```

## Interfaces

| Serviço       | URL                      |
|---------------|--------------------------|
| API Pedidos   | http://localhost:3000    |
| Kafka UI      | http://localhost:8080    |
| Dozzle        | http://localhost:9999    |
