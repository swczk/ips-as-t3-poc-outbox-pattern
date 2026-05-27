import { Kafka } from 'kafkajs';
import { pool } from '../db';
import { log } from '../shared/logger';

const KAFKA_BROKERS  = (process.env.KAFKA_BROKERS ?? 'localhost:9092').split(',');
const TOPIC          = process.env.PEDIDOS_EVENTS_TOPIC ?? 'pedidos.events';
const GROUP_ID       = process.env.CONSUMER_GROUP_ID ?? 'pedidos-consumer-group';
const KAFKA_RETRIES  = parseInt(process.env.KAFKA_RETRY_COUNT ?? '5', 10);

async function iniciar(): Promise<void> {
  const kafka = new Kafka({ clientId: 'pedidos-consumer', brokers: KAFKA_BROKERS, retry: { retries: KAFKA_RETRIES } });
  const consumer = kafka.consumer({ groupId: GROUP_ID });
  await consumer.connect();
  await consumer.subscribe({ topic: TOPIC, fromBeginning: true });

  const shutdown = async (signal: string): Promise<void> => {
    log('info', `${signal} recebido — a desligar consumer...`);
    await consumer.disconnect();
    await pool.end();
    process.exit(0);
  };

  process.on('SIGTERM', () => void shutdown('SIGTERM'));
  process.on('SIGINT',  () => void shutdown('SIGINT'));

  log('info', 'Consumer iniciado', { topic: TOPIC, groupId: GROUP_ID });

  await consumer.run({
    autoCommit: false,
    eachMessage: async ({ topic, partition, message }) => {
      const raw = message.value?.toString();
      if (!raw) return;

      let evento: { id: string; tipo: string };
      try {
        evento = JSON.parse(raw);
      } catch {
        log('warn', 'Mensagem inválida (não é JSON)', { offset: message.offset });
        await consumer.commitOffsets([{ topic, partition, offset: (BigInt(message.offset) + 1n).toString() }]);
        return;
      }

      const { rowCount } = await pool.query(
        'INSERT INTO eventos_processados (evento_id) VALUES ($1) ON CONFLICT DO NOTHING',
        [evento.id],
      );

      if (rowCount === 0) {
        log('warn', 'Evento duplicado ignorado', { id: evento.id });
      } else {
        log('info', 'Evento processado', { id: evento.id, tipo: evento.tipo });
      }

      await consumer.commitOffsets([{ topic, partition, offset: (BigInt(message.offset) + 1n).toString() }]);
    },
  });
}

iniciar().catch((err) => {
  log('error', 'Erro fatal no consumer', { erro: err instanceof Error ? err.message : String(err) });
  process.exit(1);
});
