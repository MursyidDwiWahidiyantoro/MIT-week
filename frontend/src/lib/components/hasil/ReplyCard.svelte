<script lang="ts">
  import Button from '$lib/components/ui/Button.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import templates from '$lib/data/reply-templates.json';
  import { isMockMode, API_BASE } from '$lib/api/config';
  import { isMockToken } from '$lib/api/mock';
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
      const res = await fetch(`${API_BASE}/reply`, {
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
        class="chip"
        class:selected={tone === option.value}
        aria-pressed={tone === option.value}
        onclick={() => {
          tone = option.value;
          request();
        }}
      >
        {option.label}
      </button>
    {/each}
  </div>

  <label class="field">
    <span>Draf balasan {loading ? '(menyusun…)' : ''}</span>
    <textarea class="textarea" rows="4" bind:value={text} maxlength="280"></textarea>
  </label>

  <div class="actions">
    <Button variant="quiet" onclick={copy}>
      <Icon name={copied ? 'check' : 'copy'} size={18} />
      {copied ? 'Tersalin' : 'Salin'}
    </Button>
    <Button variant="quiet" onclick={whatsapp}>
      <Icon name="whatsapp" size={18} /> Kirim lewat WhatsApp
    </Button>
    <span class="src">
      {source === 'llm' ? 'Draf AI, sudah dicek aturan sopan dan tautannya' : 'Templat bawaan AMBA'}
    </span>
  </div>
</div>

<style>
  .reply { display: grid; gap: var(--space-4); }
  .hint { margin: 0; color: var(--body-2); font-size: var(--text-sm); max-width: var(--measure); }
  .tones { display: flex; gap: var(--space-2); flex-wrap: wrap; }
  .chip {
    min-height: 44px;
    padding: 0 18px;
    background: var(--white);
    color: var(--ink);
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius-pill);
    font: inherit;
    font-weight: 700;
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out);
  }
  .chip:hover { background: var(--indigo-tint); }
  .chip.selected { background: var(--brand); color: var(--on-brand); }
  .chip.selected:hover { background: var(--brand); color: var(--on-brand); }
  .chip:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .field { display: grid; gap: var(--space-2); }
  .field span { font-weight: 700; }
  .actions { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; }
  .src { color: var(--muted); font-size: var(--text-sm); }
</style>
