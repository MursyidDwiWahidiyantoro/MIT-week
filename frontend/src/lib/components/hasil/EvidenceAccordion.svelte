<script lang="ts">
  import type { DossierParsed } from '$lib/api/schemas';
  import Skeleton from '$lib/components/ui/Skeleton.svelte';
  import { ChevronDown } from '@lucide/svelte';

  interface Props {
    pillars: DossierParsed['pillars'];
    pending?: boolean;
  }

  let { pillars, pending = false }: Props = $props();

  const LABELS = {
    ela: 'Analisis gambar (ELA)',
    ai: 'Deteksi jejak AI',
    fact: 'Pengecekan fakta',
    link: 'Pemeriksaan tautan'
  } as const;

  type Key = keyof typeof LABELS;
  const order: Key[] = ['ela', 'ai', 'fact', 'link'];

  const available = (k: Key) => pillars[k].available;

  // The pillar union has no shared `score`; narrow by key instead of casting.
  function pillarScore(p: DossierParsed['pillars'], k: Key): number {
    switch (k) {
      case 'ela':
        return p.ela.score;
      case 'ai':
        return p.ai.score;
      case 'fact':
        return p.fact.status === 'CONTRADICTS' ? 100 : p.fact.status === 'UNVERIFIED' ? 40 : 0;
      case 'link':
        return p.link.score;
    }
  }
</script>

<div class="accordion">
  {#each order as key (key)}
    <details class="item" open={available(key)}>
      <summary>
        <span class="name">{LABELS[key]}</span>
        {#if pending}
          <span class="state pending">memeriksa…</span>
        {:else if !available(key)}
          <span class="state skipped">Tidak dapat diperiksa</span>
        {:else}
          <span class="state done">skor {pillarScore(pillars, key)}</span>
        {/if}
        <span class="chev" aria-hidden="true"><ChevronDown size={16} /></span>
      </summary>

      <div class="body">
        {#if pending}
          <Skeleton lines={2} />
        {:else if !available(key)}
          <p class="note">Pemeriksaan ini tidak berjalan untuk konten ini. Hasil di atas tidak memakainya.</p>
        {:else if key === 'ela'}
          <p>
            Keandalan {Math.round(pillars.ela.reliability * 100)}%.
            {#if pillars.ela.reliability_note}{pillars.ela.reliability_note}{/if}
          </p>
          {#if pillars.ela.anomalies.length}
            <p>{pillars.ela.anomalies.length} area perlu dilihat.</p>
          {:else}
            <p>Tidak ada area menonjol yang terdeteksi.</p>
          {/if}
        {:else if key === 'ai'}
          <p>{pillars.ai.caveat}</p>
          {#if (pillars.ai.indicators ?? []).length}
            <ul>
              {#each pillars.ai.indicators ?? [] as indicator (indicator)}<li>{indicator}</li>{/each}
            </ul>
          {/if}
        {:else if key === 'fact'}
          {#if pillars.fact.claim}<p>Klaim: “{pillars.fact.claim}”</p>{/if}
          <p>Status: {pillars.fact.status}</p>
          {#if (pillars.fact.sources ?? []).length}
            <ul>
              {#each pillars.fact.sources ?? [] as s (s.url)}
                <li>
                  {#if s.match !== 'UNVERIFIED'}
                    <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- external publisher URL, not an app route -->
                    <a href={s.url} rel="noopener noreferrer nofollow" target="_blank">{s.title || s.publisher}</a>
                  {:else}
                    <span>{s.title || s.publisher}</span>
                  {/if}
                  <span class="muted"> — {s.publisher}</span>
                </li>
              {/each}
            </ul>
          {:else}
            <p class="note">Tidak ada sumber terverifikasi untuk klaim ini.</p>
          {/if}
        {:else if key === 'link'}
          {#if pillars.link.final_url}<p>Alamat akhir: {pillars.link.final_url}</p>{/if}
          {#if (pillars.link.reasons ?? []).length}
            <ul>
              {#each pillars.link.reasons ?? [] as r (r.code)}<li>{r.text}</li>{/each}
            </ul>
          {:else}
            <p>Tidak ada tanda mencurigakan pada tautan.</p>
          {/if}
        {/if}
      </div>
    </details>
  {/each}
</div>

<style>
  .accordion { display: grid; gap: var(--space-2); }
  .item { border: var(--border-width) solid var(--border); border-radius: var(--radius-md); background: var(--surface); }
  summary {
    display: flex; align-items: center; gap: var(--space-3);
    min-height: 44px; padding: var(--space-3) var(--space-4);
    cursor: pointer; list-style: none;
  }
  summary::-webkit-details-marker { display: none; }
  summary:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }
  .name { font-weight: 600; }
  .state { margin-left: auto; font-family: var(--font-mono); font-size: var(--text-xs); }
  .state.done { color: var(--text-2); }
  .state.skipped { color: var(--text-3); }
  .state.pending { color: var(--accent); }
  .chev { display: inline-flex; color: var(--text-3); transition: transform var(--dur-fast) var(--ease-out); }
  details[open] .chev { transform: rotate(180deg); }
  .body { padding: 0 var(--space-4) var(--space-4); color: var(--text-2); font-size: var(--text-sm); }
  .body p { margin: 0 0 var(--space-2); }
  .body ul { margin: 0; padding-left: var(--space-5); display: grid; gap: var(--space-1); }
  .note { color: var(--text-3); }
  .muted { color: var(--text-3); }
</style>
