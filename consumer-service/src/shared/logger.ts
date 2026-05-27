type Level = 'info' | 'warn' | 'error';

export function log(nivel: Level, msg: string, extra?: object): void {
  const entry = { nivel, ts: new Date().toISOString(), msg, ...extra };
  if (nivel === 'error') {
    console.error(JSON.stringify(entry));
  } else if (nivel === 'warn') {
    console.warn(JSON.stringify(entry));
  } else {
    console.log(JSON.stringify(entry));
  }
}
