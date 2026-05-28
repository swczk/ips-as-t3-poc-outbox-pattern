type Level = 'info' | 'warn' | 'error';

export function log(level: Level, msg: string, extra?: object): void {
  const entry = { time: new Date().toISOString(), level: level.toUpperCase(), msg, ...extra };
  if (level === 'error') {
    console.error(JSON.stringify(entry));
  } else if (level === 'warn') {
    console.warn(JSON.stringify(entry));
  } else {
    console.log(JSON.stringify(entry));
  }
}
