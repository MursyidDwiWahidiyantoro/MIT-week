<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import templates from '$lib/data/reply-templates.json';
  import { isMockMode } from '$lib/api/config';
  import { isMockToken } from '$lib/api/mock';
  import { Copy, Check, MessageCircle } from '@lucide/svelte';
  import type { Status } from '$lib/api/types';

  interface Props {
    token: string;
    status: Status;
    /** Bundled demo dossiers have no signed token, so never call the API. */
    demo?: boolean;
  }

  let { token, status, demo = false }: Props = $props();

  type Tone = 'hormat' | 'santai' | 'singkat';
  const TONES: Array<{ value: Tone; label: string }> = [
    { value: 'hormat', label: 'Hormat' },
    { value: 'santai', label: 'Santai' },
    { value: 'singkat', label: 'Singkat' }
  ];

  let tone = $state<Tone>('hormat');
  let text = $state('');
  let source = $state<'template' | 'llm'>('template');
  let loading = $state(false);
  let copied = $state(false);

  // The template renders instantly; the network draft only replaces it if it
  // validates. Offline, the template is the answer.
  const templateText = $derived(
    (templates.templates as Record<string, Record<Tone, string>>)[status]?.[tone] ?? ''
  );

  $effect(() => {
    text = templateText;
    source = 'template';
  });

  async function request(): Promise<void> {
    // Demo and preview results have no server to call. The bundled template is
    // the honest answer, and it already meets the length and tone rules.
    if (demo || isMockMode || isMockToken(token)) {
      text = templateText;
      source = 'template';
      return;
    }

    loading = true;
    try {
      const res = await fetch(`${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/reply`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, tone })
      });
      if (res.ok) {
        const body = (await res.json()) as { text?: string; source?: 'llm' | 'template' };
        if (body.text && body.text.length <= 280) {
          text = body.text;
          source = body.source ?? 'template';
        }
      }
    } catch {
      // Keep the template: it is already on screen and valid offline.
    } finally {
      loading = false;
    }
  }

  async function copy(): Promise<void> {
    try {
      await navigator.clipboard.writeText(text);
      copied = true;
      setTimeout(() => (copied = false), 2500);
    } catch {
      // Manual selection remains possible in the textarea.
    }
  }

  function whatsapp(): void {
    window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank', 'noopener');
  }
</script>

<div class="reply">
  <p class="hint">
    Balasan ini membahas isi pesannya, bukan orangnya. Tidak ada tuduhan dan tidak ada klaim pasti.
  </p>

  <div class="tones" role="group" aria-label="Pilih nada balasan">
    {#each TONES as option (option.value)}
      <button
        type="button"
        class:selected={tone === option.value}
        aria-pressed={tone === option.value}
        onclick={() => { tone = option.value; request(); }}
      >
        {option.label}
      </button>
    {/each}
  </div>

  <label class="field">
    <span>Draf balasan {loading ? '(menyusun…)' : ''}</span>
    <textarea rows="4" bind:value={text} maxlength="280"></textarea>
  </label>

  <div class="actions">
    <Button variant="quiet" onclick={copy}>
      {#if copied}<Check size={16} aria-hidden="true" /> Tersalin{:else}<Copy size={16} aria-hidden="true" /> Salin{/if}
    </Button>
    <Button variant="quiet" onclick={whatsapp}>
      <MessageCircle size={16} aria-hidden="true" /> Kirim lewat WhatsApp
    </Button>
    <span class="src">Sumber: {source === 'llm' ? 'draf model' : 'templat'}</span>
  </div>
</div>

<style>
  .reply { display: grid; gap: var(--space-3); }
  .hint { margin: 0; color: var(--text-2); font-size: var(--text-sm); max-width: var(--measure); }
  .tones { display: flex; gap: var(--space-2); flex-wrap: wrap; }
  .tones button {
    min-height: 36px; padding: 0 var(--space-3);
    background: var(--surface-2); color: var(--text-2);
    border: var(--border-width) solid var(--border); border-radius: var(--radius-md);
    font: inherit; font-size: var(--text-sm); cursor: pointer;
  }
  .tones button.selected { background: var(--selected-bg); color: var(--text); border-color: var(--border-strong); font-weight: 600; }
  .tones button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .field { display: grid; gap: var(--space-1); }
  .field span { font-size: var(--text-sm); color: var(--text-2); }
  textarea {
    padding: var(--space-3); background: var(--surface-2); color: var(--text);
    border: var(--border-width) solid var(--border-strong); border-radius: var(--radius-md);
    font: inherit; resize: vertical;
  }
  textarea:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .actions { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; }
  .src { color: var(--text-3); font-size: var(--text-xs); font-family: var(--font-mono); }
</style>
