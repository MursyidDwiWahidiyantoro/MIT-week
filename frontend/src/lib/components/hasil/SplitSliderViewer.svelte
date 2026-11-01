<script lang="ts">
  import type { Anomaly } from '$lib/api/types';

  interface Props {
    src: string;
    elaSrc?: string | null;
    anomalies?: Anomaly[];
    alt?: string;
  }

  let { src, elaSrc = null, anomalies = [], alt = 'Gambar yang diperiksa' }: Props = $props();

  let position = $state(50);
  let view = $state<'slider' | 'original' | 'ela'>('slider');

  const hasEla = $derived(!!elaSrc);

  function onRangeKey(e: KeyboardEvent) {
    const step = e.key === 'PageUp' || e.key === 'PageDown' ? 10 : 2;
    if (e.key === 'ArrowRight' || e.key === 'PageUp') {
      position = Math.min(100, position + step);
      e.preventDefault();
    } else if (e.key === 'ArrowLeft' || e.key === 'PageDown') {
      position = Math.max(0, position - step);
      e.preventDefault();
    } else if (e.key === 'Home') {
      position = 0;
      e.preventDefault();
    } else if (e.key === 'End') {
      position = 100;
      e.preventDefault();
    }
  }

  // Boxes are normalized 0-1000 as [ymin, xmin, ymax, xmax] against the image's
  // own dimensions, so percentages hold at any aspect ratio.
  const boxes = $derived(
    anomalies.map((a) => {
      const [ymin, xmin, ymax, xmax] = a.bounding_box;
      return {
        top: ymin / 10,
        left: xmin / 10,
        width: (xmax - xmin) / 10,
        height: (ymax - ymin) / 10,
        reason: a.reason ?? '',
        text: a.extracted_text ?? ''
      };
    })
  );
</script>

<div class="viewer">
  <div class="tabs" role="tablist" aria-label="Cara melihat gambar">
    {#if hasEla}
      <button type="button" role="tab" aria-selected={view === 'slider'} onclick={() => (view = 'slider')}>Geser</button>
    {/if}
    <button type="button" role="tab" aria-selected={view === 'original'} onclick={() => (view = 'original')}>Asli</button>
    {#if hasEla}
      <button type="button" role="tab" aria-selected={view === 'ela'} onclick={() => (view = 'ela')}>ELA</button>
    {/if}
  </div>

  <div class="stage">
    <img {src} {alt} class="base" />

    {#if hasEla && view !== 'original'}
      <div class="overlay" style:clip-path={view === 'slider' ? `inset(0 0 0 ${position}%)` : 'none'}>
        <img src={elaSrc} alt="Peta perbedaan kompresi (ELA)" class="base fill" />
      </div>
    {/if}

    <!-- Boxes sit OUTSIDE the clip layer so they stay visible on both sides. -->
    <div class="boxes" aria-hidden="true">
      {#each boxes as box, i (i)}
        <span
          class="box"
          style:top="{box.top}%"
          style:left="{box.left}%"
          style:width="{box.width}%"
          style:height="{box.height}%"
        ></span>
      {/each}
    </div>

    {#if hasEla && view === 'slider'}
      <div class="handle" style:left="{position}%" aria-hidden="true"></div>
    {/if}
  </div>

  {#if hasEla && view === 'slider'}
    <input
      type="range"
      min="0"
      max="100"
      step="1"
      bind:value={position}
      aria-label="Geser untuk membandingkan gambar asli dan hasil ELA"
      aria-valuetext="{position}% gambar asli"
      onkeydown={onRangeKey}
    />
  {/if}
</div>

{#if boxes.length > 0}
  <ul class="anomalies">
    {#each boxes as box, i (i)}
      <li>
        <strong>Area {i + 1}</strong>
        {#if box.text}<span class="quote">“{box.text}”</span>{/if}
        {#if box.reason}<span class="reason">{box.reason}</span>{/if}
      </li>
    {/each}
  </ul>
{/if}

<style>
  .viewer { display: grid; gap: var(--space-3); }
  .tabs {
    display: inline-flex; gap: var(--space-1); padding: var(--space-1);
    border: var(--border-width) solid var(--border); border-radius: var(--radius-md);
    justify-self: start;
  }
  .tabs button {
    min-height: 36px; padding: 0 var(--space-3); background: transparent; color: var(--text-2);
    border: 0; border-radius: var(--radius-sm); font: inherit; font-size: var(--text-sm); cursor: pointer;
  }
  .tabs button[aria-selected='true'] { background: var(--selected-bg); color: var(--text); font-weight: 600; }
  .tabs button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }

  .stage {
    position: relative; line-height: 0;
    border-radius: var(--radius-md); overflow: hidden; background: var(--surface-2);
  }
  .base { width: 100%; height: auto; display: block; }
  .fill { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: fill; }
  .overlay { position: absolute; inset: 0; }

  .boxes { position: absolute; inset: 0; pointer-events: none; }
  .box {
    position: absolute;
    border: 2px solid var(--err);
    border-radius: 2px;
    box-shadow: 0 0 0 1px rgb(0 0 0 / 0.4);
  }
  .handle { position: absolute; top: 0; bottom: 0; width: 2px; background: var(--accent); transform: translateX(-1px); }

  input[type='range'] { width: 100%; min-height: 44px; accent-color: var(--accent); }
  input[type='range']:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }

  .anomalies { list-style: none; margin: var(--space-3) 0 0; padding: 0; display: grid; gap: var(--space-2); }
  .anomalies li { display: flex; flex-wrap: wrap; gap: var(--space-2); font-size: var(--text-sm); color: var(--text-2); }
  .quote { color: var(--text); }
</style>
