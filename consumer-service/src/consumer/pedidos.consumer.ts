import { Kafka } from 'kafkajs';
import Redis from 'ioredis';
import { log } from '../shared/logger';
import { handleMessage } from './message.handler';

const KAFKA_BROKERS = (process.env.KAFKA_BROKERS ?? 'localhost:9092').split(',');
const TOPIC         = process.env.PEDIDOS_EVENTS_TOPIC ?? 'pedidos.events';
const GROUP_ID      = process.env.CONSUMER_GROUP_ID ?? 'pedidos-consumer-group';
const REDIS_URL     = process.env.REDIS_URL ?? 'redis://localhost:6379';
const KAFKA_RETRIES = parseInt(process.env.KAFKA_RETRY_COUNT ?? '5', 10);

async function iniciar(): Promise<void> {
  const redis = new Redis(REDIS_URL);

  const kafka = new Kafka({
    clientId: 'pedidos-consumer',
    brokers: KAFKA_BROKERS,
    retry: { retries: KAFKA_RETRIES },
  });

  const consumer = kafka.consumer({ groupId: GROUP_ID });
  await consumer.connect();

  await consumer.subscribe({ topic: TOPIC, fromBeginning: true });

  const shutdown = async (): Promise<void> => {
    log('info', 'consumer shutting down');
    await consumer.disconnect();
    redis.disconnect();
    process.exit(0);
  };

  process.on('SIGTERM', () => void shutdown());
  process.on('SIGINT',  () => void shutdown());

  log('info', 'consumer started', { topic: TOPIC, groupId: GROUP_ID, brokers: KAFKA_BROKERS });

  await consumer.run({
    autoCommit: false,
    eachMessage: async ({ topic, partition, message }) => {
      try {
        await handleMessage(redis, topic, GROUP_ID, partition, message.offset, message.key, message.value);
      } catch (err) {
        log('error', 'failed to process event', {
          eventId: message.key?.toString() ?? message.offset,
          error: err instanceof Error ? err.message : String(err),
        });
        throw err;
      }

      await consumer.commitOffsets([
        { topic, partition, offset: (BigInt(message.offset) + 1n).toString() },
      ]);
    },
  });
}

iniciar().catch((err) => {
  log('error', 'consumer fatal error', {
    error: err instanceof Error ? err.message : String(err),
  });
  process.exit(1);
});
