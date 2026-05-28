import Redis from 'ioredis';
import { log } from '../shared/logger';

interface PedidoCriadoPayload {
  pedidoId: string;
  clienteId: string;
  valor: string;
  timestamp: string;
}

export async function handleMessage(
  redis: Redis,
  topic: string,
  groupId: string,
  partition: number,
  offset: string,
  key: Buffer | null,
  rawValue: Buffer | null,
): Promise<void> {
  const raw = rawValue?.toString();

  if (!raw) return;

  let payload: PedidoCriadoPayload;
  try {
    payload = JSON.parse(raw) as PedidoCriadoPayload;
  } catch {
    log('error', 'invalid payload — message discarded', {
      topic,
      partition,
      offset,
      raw: raw.slice(0, 200),
    });
    return;
  }

  const chave = `idempotencia:${payload.pedidoId}`;
  const resultado = await redis.set(chave, '1', 'EX', 86400, 'NX');

  if (resultado === null) {
    log('warn', 'duplicate event skipped', {
      eventId: payload.pedidoId,
    });
    return;
  }

  log('info', 'event processed', {
    eventId: key?.toString() ?? offset,
    pedidoId: payload.pedidoId,
    topic,
    groupId,
  });
}
