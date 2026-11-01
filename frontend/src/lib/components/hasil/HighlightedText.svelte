<script lang="ts">
  import type { Tactic } from '$lib/api/types';
  import { Info } from '@lucide/svelte';

  interface Props {
    text: string;
    tactics: Tactic[];
    labels: Record<string, { label: string; why: string }>;
  }

  let { text, tactics, labels }: Props = $props();

  /**
   * Splits text into plain and marked segments. Matching is case-insensitive
   * and index-based; every piece renders through Svelte interpolation, so no
   * user or model content is ever injected as HTML.
   */
  const segments = $derived.by(() => {
    type Info = { label: string; why: string };
    type Mark = { start: number; end: number; info: Info };
    const marks: Mark[] = [];
    const haystack = text.toLowerCase();

    for (const tactic of tactics) {
      const info = labels[tactic.code];
      if (!info) continue;
      const needle = tactic.phrase.toLowerCase();
      if (!needle) continue;
      let from = 0;
      while (from < haystack.length) {
        const idx = haystack.indexOf(needle, from);
        if (idx === -1) break;
        if (!marks.some((m) => idx < m.end && idx + needle.length > m.start)) {
          marks.push({ start: idx, end: idx + needle.length, info });
        }
        from = idx + needle.length;
      }
    }

    marks.sort((a, b) => a.start - b.start);
    const out: Array<{ text: string; tactic?: Info }> = [];
    let cursor = 0;
    for (const m of marks) {
      if (m.start > cursor) out.push({ text: text.slice(cursor, m.start) });
      out.push({ text: text.slice(m.start, m.end), tactic: m.info });
      cursor = m.end;
    }
    if (cursor < text.length) out.push({ text: text.slice(cursor) });
    return out;
  });
</script>

<p class="text">
  {#each segments as seg, i (i)}{#if seg.tactic}<mark class="tactic">{seg.text}<span class="tag">
        <span class="tag-icon" aria-hidden="true"><Info size={12} /></span>
        {seg.tactic.label}
        <span class="pop" role="note">{seg.tactic.why}</span>
      </span></mark>{:else}{seg.text}{/if}{/each}
</p>

<style>
  .text { white-space: pre-wrap; line-height: var(--leading-normal); margin: 0; }
  .tactic {
    background: color-mix(in srgb, var(--warn) 18%, transparent);
    color: var(--text);
    border-radius: 2px;
    padding: 0 2px;
  }
  .tag {
    display: inline-flex; align-items: center; gap: 4px;
    margin-left: var(--space-2);
    padding: 1px 6px;
    border: var(--border-width) solid var(--warn);
    border-radius: var(--radius-sm);
    font-size: var(--text-xs);
    color: var(--warn);
    position: relative;
  }
  .tag-icon { display: inline-flex; }
  .pop {
    position: absolute; left: 0; top: 100%; z-index: 10;
    width: max-content; max-width: 18rem;
    margin-top: var(--space-1);
    padding: var(--space-2) var(--space-3);
    background: var(--surface-2);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-sm);
    color: var(--text-2);
    font-size: var(--text-xs);
    line-height: var(--leading-normal);
    display: none;
  }
  .tag:hover .pop, .tag:focus-within .pop { display: block; }
</style>
