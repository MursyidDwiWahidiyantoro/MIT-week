<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import Quiz from '$lib/components/belajar/Quiz.svelte';
  import forensicQuiz from '$lib/data/forensic-quiz.json';
  import {
    AI_CUES,
    HOAX_CUES,
    CHECK_STEPS,
    SCAMS,
    RULES,
    SAFE_WORDS,
    UI,
    type Text
  } from '$lib/data/belajar';

  const tx = (text: Text): string => text.id;

  const NAV = [
    { id: 'video-ai', label: UI.nav.ai },
    { id: 'video-hoaks', label: UI.nav.hoax },
    { id: 'cek-sendiri', label: UI.nav.check },
    { id: 'penipuan', label: UI.nav.scam },
    { id: 'kuis', label: UI.nav.quiz },
    { id: 'keluarga', label: UI.nav.family }
  ] as const;

  // ---- ELA practice (PRD F7): three synthetic images with the heat map AMBA computed. ----
  type Guess = 'asli' | 'diedit' | 'belum-pasti';
  const GUESSES: Guess[] = ['asli', 'diedit', 'belum-pasti'];
  const forensic = forensicQuiz.items as Array<{ id: string; image: string; ela: string; answer: Guess }>;
  let guesses = $state<Record<string, Guess | undefined>>({});
  let revealed = $state<Record<string, boolean>>({});

  // ---- family safe word: made in the browser, never sent anywhere ----
  let safeWord = $state('');
  let copied = $state(false);

  function generate(): void {
    const buf = new Uint32Array(2);
    crypto.getRandomValues(buf);
    const a = SAFE_WORDS[buf[0] % SAFE_WORDS.length];
    const b = SAFE_WORDS[buf[1] % SAFE_WORDS.length];
    safeWord = a === b ? `${a} Dua` : `${a} ${b}`;
    copied = false;
  }

  async function copy(): Promise<void> {
    try {
      await navigator.clipboard.writeText(safeWord);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // Selecting the text by hand still works.
    }
  }
</script>

<svelte:head>
  <title>{tx(UI.title)}</title>
  <meta name="description" content={tx(UI.description)} />
</svelte:head>

<div class="belajar">
  <header class="head">
    <h1 class="h1">{tx(UI.h1)}</h1>
    <p class="lede">{tx(UI.lede)}</p>
    <nav class="jump" aria-label={tx(UI.jump)}>
      <ul>
        {#each NAV as item (item.id)}
          <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- same-page anchor -->
          <li><a href={`#${item.id}`}>{tx(item.label)}</a></li>
        {/each}
      </ul>
    </nav>
  </header>

  <section id="video-ai" aria-labelledby="ai-title">
    <h2 class="h2" id="ai-title">{tx(UI.ai.title)}</h2>
    <p class="sec-lede">{tx(UI.ai.lede)}</p>
    <ul class="cards cards--3">
      {#each AI_CUES as c (c.id)}
        <li class="tile"><h3>{tx(c.title)}</h3><p>{tx(c.body)}</p></li>
      {/each}
    </ul>
    <p class="note">{tx(UI.ai.note)}</p>
  </section>

  <section id="video-hoaks" aria-labelledby="hoax-title">
    <h2 class="h2" id="hoax-title">{tx(UI.hoax.title)}</h2>
    <p class="sec-lede">{tx(UI.hoax.lede)}</p>
    <ul class="cards cards--2">
      {#each HOAX_CUES as c (c.id)}
        <li class="tile"><h3>{tx(c.title)}</h3><p>{tx(c.body)}</p></li>
      {/each}
    </ul>
  </section>

  <section id="cek-sendiri" aria-labelledby="check-title">
    <h2 class="h2" id="check-title">{tx(UI.check.title)}</h2>
    <p class="sec-lede">{tx(UI.check.lede)}</p>
    <ol class="steps">
      {#each CHECK_STEPS as step, i (step.id)}
        <li class="step">
          <span class="num" aria-hidden="true">{i + 1}</span>
          <div>
            <h3>{tx(step.title)}</h3>
            <p>{tx(step.body)}</p>
            {#if step.link}
              <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- external site, not an app route -->
              <a class="ext" href={step.link.href} target="_blank" rel="noopener noreferrer">
                {step.link.label}
                <span class="sr-only"> ({tx(UI.check.newTab)})</span>
              </a>
            {/if}
          </div>
        </li>
      {/each}
    </ol>
  </section>

  <section id="penipuan" aria-labelledby="scam-title">
    <h2 class="h2" id="scam-title">{tx(UI.scam.title)}</h2>
    <p class="sec-lede">{tx(UI.scam.lede)}</p>
    <ul class="cards cards--2">
      {#each SCAMS as c (c.id)}
        <li class="tile">
          <h3>{tx(c.title)}</h3>
          <p>{tx(c.body)}</p>
          <p class="act"><strong>{tx(UI.scam.act)}:</strong> {tx(c.act)}</p>
        </li>
      {/each}
    </ul>
    <p class="rule"><Icon name="lock" size={20} /> <span>{tx(UI.scam.rule)}</span></p>
  </section>

  <section id="kuis" aria-labelledby="quiz-title">
    <h2 class="h2" id="quiz-title">{tx(UI.quiz.title)}</h2>
    <p class="sec-lede">{tx(UI.quiz.lede)}</p>
    <Quiz />
  </section>

  <section class="forensic" aria-labelledby="forensic-title">
    <h2 class="h2" id="forensic-title">{t('fq.title')}</h2>
    <p class="sec-lede">{t('fq.lede')}</p>
    <div class="forensic-grid">
      {#each forensic as item, i (item.id)}
        <article class="f-item tile">
          <img class="f-img" src={item.image} alt={t('fq.imageAlt', { n: i + 1 })} width="360" height="480" loading="lazy" />
          <div class="f-body">
            <p class="f-legend" id={`fq-${item.id}`}>{t('fq.question')}</p>
            <div class="f-options" role="radiogroup" aria-labelledby={`fq-${item.id}`}>
              {#each GUESSES as g (g)}
                <label class="option">
                  <input type="radio" name={`fq-${item.id}`} value={g} bind:group={guesses[item.id]} />
                  <span>{t(`fq.opt.${g}`)}</span>
                </label>
              {/each}
            </div>
            <button
              class="btn btn--quiet btn--sm"
              type="button"
              aria-expanded={!!revealed[item.id]}
              aria-controls={`fq-panel-${item.id}`}
              disabled={!guesses[item.id]}
              onclick={() => (revealed = { ...revealed, [item.id]: !revealed[item.id] })}
            >
              <Icon name="image" size={16} /> {revealed[item.id] ? t('fq.hide') : t('fq.reveal')}
            </button>
            {#if revealed[item.id]}
              <div class="f-reveal" id={`fq-panel-${item.id}`}>
                <figure>
                  <img src={item.ela} alt={t('fq.elaAlt', { n: i + 1 })} width="360" height="480" />
                  <figcaption>{t('fq.ela')}</figcaption>
                </figure>
                <p class="f-verdict" role="status">
                  {guesses[item.id] === item.answer ? t('fq.right') : t('fq.wrong', { answer: t(`fq.opt.${item.answer}`) })}
                </p>
                <p class="f-explain">{t(`fq.explain.${item.answer}`)}</p>
              </div>
            {/if}
          </div>
        </article>
      {/each}
    </div>
  </section>

  <section id="keluarga" class="family" aria-labelledby="family-title">
    <div class="no-print">
      <h2 class="h2" id="family-title">{tx(UI.family.title)}</h2>
      <p class="sec-lede">{tx(UI.family.lede)}</p>
      <p class="how">{tx(UI.family.how)}</p>

      <div class="generator">
        <output class="word" aria-live="polite">{safeWord || tx(UI.family.none)}</output>
        <div class="gen-actions">
          <button class="btn btn--primary" type="button" onclick={generate}>
            <Icon name="refresh" size={18} /> {safeWord ? tx(UI.family.again) : tx(UI.family.make)}
          </button>
          {#if safeWord}
            <button class="btn btn--quiet" type="button" onclick={copy}>
              <Icon name={copied ? 'check' : 'copy'} size={18} /> {copied ? tx(UI.family.copied) : tx(UI.family.copy)}
            </button>
          {/if}
        </div>
      </div>
      <p class="note">{tx(UI.family.note)}</p>

      <h3 class="rules-title">{tx(UI.family.rules)}</h3>
      <ol class="cards cards--3 rules">
        {#each RULES as r (r.id)}
          <li class="tile"><h3>{tx(r.title)}</h3><p>{tx(r.body)}</p></li>
        {/each}
      </ol>
    </div>

    <div class="printable">
      <p class="p-brand">{tx(UI.family.cardBrand)}</p>
      <p class="p-label">{tx(UI.family.cardLabel)}</p>
      <p class="p-word">{safeWord || '____________________'}</p>
      <ol class="p-rules">
        {#each RULES as r (r.id)}<li>{tx(r.title)}</li>{/each}
      </ol>
      <p class="p-foot">{tx(UI.family.cardFoot)}</p>
    </div>
    <button class="btn btn--quiet no-print" type="button" onclick={() => window.print()}>
      <Icon name="print" size={18} /> {tx(UI.family.print)}
    </button>
  </section>
</div>

<style>
  .belajar { display: grid; gap: var(--space-8); }
  .head { max-width: 54rem; }
  .head .lede { margin-top: var(--space-3); }

  .jump ul { display: flex; flex-wrap: wrap; gap: var(--space-2); list-style: none; margin: var(--space-5) 0 0; padding: 0; }
  .jump a {
    display: inline-flex;
    align-items: center;
    min-height: 44px;
    padding: 0 var(--space-4);
    border: 1px solid var(--glass-edge);
    border-radius: var(--radius-pill);
    background: var(--glass-bg);
    box-shadow: var(--glass-highlight);
    font-weight: 700;
    font-size: var(--text-sm);
    text-decoration: none;
    color: var(--ink);
  }
  .jump a:hover { background: var(--glass-bg-strong); color: var(--ink); }

  section { scroll-margin-top: 7rem; }
  .sec-lede { margin: var(--space-2) 0 var(--space-5); color: var(--body-2); max-width: var(--measure); }
  .note { margin: var(--space-4) 0 0; color: var(--body-2); font-size: var(--text-sm); max-width: 46rem; }

  /* Read, not browse: the lists sit on hairlines instead of a grid of equal cards. */
  .cards { list-style: none; margin: 0; padding: 0; display: grid; gap: 0 clamp(1.5rem, 3vw, 2.75rem); grid-template-columns: 1fr; }
  @media (min-width: 40rem) { .cards--2 { grid-template-columns: repeat(2, minmax(0, 1fr)); } .cards--3 { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  @media (min-width: 64rem) { .cards--3 { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
  .tile {
    min-width: 0;
    padding: var(--space-4) 0 var(--space-5);
    border-top: 1px solid var(--border);
  }
  .tile h3 { font-family: var(--font-sans); font-weight: 700; font-size: var(--text-base); letter-spacing: 0; margin: 0 0 var(--space-1); }
  .tile p { margin: 0; color: var(--body-2); font-size: var(--text-base); line-height: 1.5; }
  .tile .act { margin-top: var(--space-3); color: var(--ink); }

  .steps { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--space-3); max-width: 54rem; }
  .step { display: flex; gap: var(--space-4); align-items: flex-start; padding: var(--space-4) var(--space-5); background: var(--glass-bg); border: 1px solid var(--glass-edge); border-radius: var(--radius-md); box-shadow: var(--glass-highlight); }
  .step h3 { font-size: var(--text-lg); margin: 0 0 var(--space-1); }
  .step p { margin: 0 0 var(--space-2); color: var(--body-2); }
  .ext { display: inline-flex; align-items: center; min-height: 44px; font-weight: 700; color: var(--indigo-text); }
  .num {
    display: grid;
    place-items: center;
    flex: none;
    width: 36px;
    height: 36px;
    margin-bottom: var(--space-2);
    border-radius: 50%;
    background: var(--indigo-soft);
    color: var(--indigo-text);
    font-family: var(--font-display);
    font-weight: 700;
  }
  .step .num { margin: 0; }

  .rule {
    display: flex;
    gap: var(--space-3);
    align-items: flex-start;
    margin: var(--space-5) 0 0;
    padding: var(--space-4) var(--space-5);
    max-width: 54rem;
    background: var(--yellow-tint);
    border: 1px solid color-mix(in srgb, var(--warn) 35%, transparent);
    border-radius: var(--radius-md);
    font-weight: 700;
  }
  .rule :global(svg) { flex: none; margin-top: 3px; }

  /* ELA practice */
  .forensic-grid { display: grid; gap: var(--space-4); grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr)); }
  .f-item { display: grid; gap: var(--space-3); align-content: start; }
  /* The practice items are interactive: they keep a pane of their own. */
  .forensic-grid .tile { padding: var(--space-4); border: 1px solid var(--glass-edge); border-radius: var(--radius-lg); background: var(--glass-bg); box-shadow: var(--glass-highlight), var(--shadow-1); }
  .f-img, .f-reveal img { width: 100%; height: auto; border-radius: var(--radius-sm); border: 1px solid var(--border); display: block; }
  .f-body, .f-reveal { display: grid; gap: var(--space-3); }
  .f-body .f-legend { margin: 0; font-weight: 700; color: var(--ink); }
  .f-options { display: grid; gap: var(--space-2); }
  .option {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 48px;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    cursor: pointer;
  }
  .option:hover { background: var(--indigo-tint); }
  .option:has(input:checked) { border-color: var(--brand); box-shadow: inset 0 0 0 1px var(--brand); }
  .option input { width: 22px; height: 22px; accent-color: var(--brand); flex: none; }
  .f-reveal figure { margin: 0; }
  .f-reveal figcaption { font-size: var(--text-sm); color: var(--muted); margin-top: var(--space-1); }
  .f-reveal .f-verdict { font-weight: 700; color: var(--ink); }

  /* Family */
  .how { margin: 0 0 var(--space-4); color: var(--body-2); max-width: var(--measure); }
  .generator { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-4); }
  .word {
    min-height: 64px;
    display: inline-flex;
    align-items: center;
    padding: 0 var(--space-5);
    background: var(--safe);
    color: var(--on-fill);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font-family: var(--font-display);
    font-weight: 650;
    font-size: clamp(1.4rem, 3vw, 2rem);
    letter-spacing: -0.02em;
  }
  .gen-actions { display: flex; flex-wrap: wrap; gap: var(--space-3); }
  .rules-title { font-size: var(--text-xl); margin: var(--space-6) 0 var(--space-4); }
  .printable {
    margin: var(--space-6) 0 var(--space-4);
    padding: clamp(1.25rem, 3vw, 2rem);
    border: 2px dashed var(--border-strong);
    border-radius: var(--radius-lg);
    max-width: 34rem;
  }
  .p-brand { margin: 0 0 var(--space-4); font-family: var(--font-display); font-weight: 650; }
  .p-label { margin: 0; font-size: var(--text-sm); font-weight: 700; color: var(--body-2); }
  .p-word { margin: var(--space-2) 0 var(--space-4); font-family: var(--font-display); font-weight: 650; font-size: clamp(1.6rem, 4vw, 2.2rem); letter-spacing: -0.02em; }
  .p-rules { margin: 0 0 var(--space-4); padding-left: var(--space-5); display: grid; gap: var(--space-2); }
  .p-foot { margin: 0; color: var(--muted); font-size: var(--text-xs); }

  @media (max-width: 40rem) {
    .gen-actions .btn { width: 100%; }
    .step { padding: var(--space-4); }
  }

  /* Print: only the family card, black on white. */
  @media print {
    :global(header),
    :global(footer),
    .head,
    .belajar > section:not(.family),
    .no-print { display: none !important; }
    .printable { border: 2px solid #000; max-width: none; margin: 0; }
    :global(body) { background: #fff; color: #000; }
  }
</style>
