/**
 * POST-based SSE reader.
 *
 * EventSource cannot POST or carry an AbortSignal, so the audit endpoints are
 * consumed through fetch + ReadableStream. Heartbeat comments (`: ping`) are
 * ignored: they are not progress and must never reach the UI.
 */

export type SSEEvent =
  | { type: 'stage'; pillar: 'ela' | 'ai' | 'fact' | 'link'; status: 'running' | 'done' | 'skipped' | 'failed' }
  | { type: 'partial'; pillar: 'ela' | 'ai' | 'fact' | 'link'; data: unknown }
  | { type: 'verdict'; data: unknown }
  | { type: 'error'; data: unknown }
  | { type: 'done' };

/** Parses one SSE frame (already split on a blank line). */
export function parseFrame(frame: string): SSEEvent | null {
  let event = 'message';
  const dataLines: string[] = [];

  for (const rawLine of frame.split('\n')) {
    const line = rawLine.replace(/\r$/, '');
    if (line === '' || line.startsWith(':')) continue; // comment / heartbeat
    if (line.startsWith('event:')) {
      event = line.slice(6).trim();
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trimStart());
    }
  }

  if (event === 'message' && dataLines.length === 0) return null;
  const raw = dataLines.join('\n');

  switch (event) {
    case 'stage': {
      const d = safeParse(raw);
      if (!isObject(d)) return null;
      return {
        type: 'stage',
        pillar: d.pillar as 'ela' | 'ai' | 'fact' | 'link',
        status: d.status as 'running' | 'done' | 'skipped' | 'failed'
      };
    }
    case 'partial': {
      const d = safeParse(raw);
      if (!isObject(d)) return null;
      return { type: 'partial', pillar: d.pillar as 'ela' | 'ai' | 'fact' | 'link', data: d.data };
    }
    case 'verdict':
      return { type: 'verdict', data: safeParse(raw) };
    case 'error':
      return { type: 'error', data: safeParse(raw) };
    case 'done':
      return { type: 'done' };
    default:
      return null;
  }
}

/** Splits a growing buffer into complete frames, returning the remainder. */
export function splitFrames(buffer: string): { frames: string[]; rest: string } {
  const frames: string[] = [];
  let rest = buffer;
  let idx = rest.indexOf('\n\n');
  while (idx !== -1) {
    frames.push(rest.slice(0, idx));
    rest = rest.slice(idx + 2);
    idx = rest.indexOf('\n\n');
  }
  return { frames, rest };
}

/** Reads a response body as a stream of SSE events. */
export async function* readSSE(
  response: Response,
  signal: AbortSignal
): AsyncGenerator<SSEEvent, void, unknown> {
  if (!response.body) return;
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (!signal.aborted) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const split = splitFrames(buffer);
      buffer = split.rest;
      for (const frame of split.frames) {
        const event = parseFrame(frame);
        if (event) yield event;
      }
    }
    // Flush a trailing frame that arrived without the blank-line terminator.
    const trailing = parseFrame(buffer);
    if (trailing) yield trailing;
  } finally {
    reader.releaseLock();
  }
}

function safeParse(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null;
}
