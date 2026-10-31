import { describe, expect, it } from 'vitest';
import { parseFrame, splitFrames } from './sse';

describe('parseFrame', () => {
  it('parses a stage event', () => {
    const event = parseFrame('event: stage\ndata: {"pillar":"ela","status":"running"}');
    expect(event).toEqual({ type: 'stage', pillar: 'ela', status: 'running' });
  });

  it('parses a partial event with its payload', () => {
    const event = parseFrame('event: partial\ndata: {"pillar":"fact","data":{"claim":"x"}}');
    expect(event).toMatchObject({ type: 'partial', pillar: 'fact' });
  });

  it('parses verdict and done', () => {
    expect(parseFrame('event: verdict\ndata: {"schema_version":2}')).toMatchObject({ type: 'verdict' });
    expect(parseFrame('event: done\ndata: {}')).toEqual({ type: 'done' });
  });

  it('ignores heartbeat comments entirely', () => {
    expect(parseFrame(': ping')).toBeNull();
    expect(parseFrame('event: stage\ndata: {"pillar":"ai","status":"done"}\n: ping')).toMatchObject({
      type: 'stage',
      pillar: 'ai'
    });
  });

  it('returns null for an unknown event type', () => {
    expect(parseFrame('event: mystery\ndata: {}')).toBeNull();
  });

  it('returns null for malformed JSON rather than throwing', () => {
    expect(parseFrame('event: stage\ndata: {oops')).toBeNull();
  });
});

describe('splitFrames', () => {
  it('splits complete frames and keeps the remainder', () => {
    const { frames, rest } = splitFrames('a\n\nb\n\npartial');
    expect(frames).toEqual(['a', 'b']);
    expect(rest).toBe('partial');
  });

  it('returns no frames for an incomplete buffer', () => {
    const { frames, rest } = splitFrames('event: stage\ndata: {}');
    expect(frames).toEqual([]);
    expect(rest).toBe('event: stage\ndata: {}');
  });

  it('handles an empty buffer', () => {
    expect(splitFrames('')).toEqual({ frames: [], rest: '' });
  });
});
