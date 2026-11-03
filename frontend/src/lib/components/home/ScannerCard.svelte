<script lang="ts">
  import { resolve } from '$app/paths';
  import Icon from '$lib/components/ui/Icon.svelte';

  /**
   * A worked example, not a fake scan. Each tab is one real-world message
   * pattern with the three signs AMBA would explain, so a first-time visitor
   * learns the shape of a verdict before submitting anything.
   */
  interface Example {
    id: 'message' | 'link' | 'phone';
    tab: string;
    icon: 'chat' | 'link' | 'phone';
    label: string;
    text: string;
    verdict: string;
    score: number;
    tone: 'danger' | 'warn';
    flags: Array<{ lead: string; rest: string }>;
  }

  const EXAMPLES: Example[] = [
    {
      id: 'message',
      tab: 'Pesan',
      icon: 'chat',
      label: 'Pesan yang Anda terima',
      text: 'URGENT: Paket Anda tertahan di bea cukai. Bayar biaya rilis hari ini di bit-ly-parcel-id.co atau paket dikembalikan.',
      verdict: 'Kemungkinan scam',
      score: 92,
      tone: 'danger',
      flags: [
        { lead: 'Mendesak.', rest: 'Kurir resmi tidak mengancam dengan batas waktu.' },
        { lead: 'Tautan aneh.', rest: 'Alamatnya meniru merek pengiriman.' },
        { lead: 'Biaya mendadak.', rest: 'Biaya tidak pernah ditagih lewat tautan acak.' }
      ]
    },
    {
      id: 'link',
      tab: 'Tautan',
      icon: 'link',
      label: 'Tautan yang dikirim',
      text: 'https://secure-bank-login.verify-account.top/id',
      verdict: 'Tautan berbahaya',
      score: 96,
      tone: 'danger',
      flags: [
        { lead: 'Nama mirip.', rest: 'Tertulis bank, tapi domain aslinya lain.' },
        { lead: 'Akhiran asing.', rest: 'Bank jarang memakai akhiran web yang murah.' },
        { lead: 'Minta masuk.', rest: 'Jangan pernah masuk dari tautan yang tidak Anda minta.' }
      ]
    },
    {
      id: 'phone',
      tab: 'Telepon',
      icon: 'phone',
      label: 'Penelepon atau pengirim',
      text: '+62 811 0000 4421: “Halo Bu, ini anak saya yang baru. Nomor saya ganti, tolong simpan.”',
      verdict: 'Perlu diperiksa',
      score: 68,
      tone: 'warn',
      flags: [
        { lead: 'Nomor asing.', rest: 'Mengaku keluarga, tapi tidak ada di kontak Anda.' },
        { lead: 'Bisa berlanjut.', rest: 'Pembuka ini sering jadi langkah pertama minta uang.' },
        { lead: 'Langkah aman.', rest: 'Telepon nomor lama dan tanyakan hal yang hanya dia tahu.' }
      ]
    }
  ];

  let active = $state<Example['id']>('message');
  const current = $derived(EXAMPLES.find((e) => e.id === active) ?? EXAMPLES[0]);
</script>

<div class="scan-wrap">
  <p class="sticker sticker--coral" style="--r:7deg; top:-6px; right:14px">URGENT!! Segera</p>
  <p class="sticker sticker--yellow" style="--r:-6deg; bottom:-22px; left:8px; animation-delay:1s">
    Anda menang hadiah!
  </p>

  <div class="scan card" id="scan">
    <div class="head">
      <h2>Ini scam atau bukan?</h2>
      <span class="pill pill--soft">Contoh</span>
    </div>

    <div class="tabs" role="tablist" aria-label="Jenis contoh">
      {#each EXAMPLES as example (example.id)}
        <button
          type="button"
          role="tab"
          id={`tab-${example.id}`}
          aria-selected={active === example.id}
          aria-controls="scan-panel"
          tabindex={active === example.id ? 0 : -1}
          class:is-active={active === example.id}
          onclick={() => (active = example.id)}
          onkeydown={(e) => {
            if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
            e.preventDefault();
            const i = EXAMPLES.findIndex((x) => x.id === active);
            const next =
              e.key === 'ArrowRight'
                ? (i + 1) % EXAMPLES.length
                : (i - 1 + EXAMPLES.length) % EXAMPLES.length;
            active = EXAMPLES[next].id;
            (e.currentTarget.parentElement?.querySelectorAll<HTMLElement>('[role="tab"]')[
              next
            ])?.focus();
          }}
        >
          <Icon name={example.icon} size={18} />
          {example.tab}
        </button>
      {/each}
    </div>

    <div id="scan-panel" role="tabpanel" aria-labelledby={`tab-${current.id}`}>
      <p class="sample-label">{current.label}</p>
      <p class="sample">{current.text}</p>

      <a class="scan-bar" href={resolve('/deteksi')}>
        <Icon name="sparkle" size={18} />
        Periksa punya Anda
      </a>

      <div class="verdict" class:verdict--warn={current.tone === 'warn'}>
        <div class="verdict-top">
          <span class="chip chip--{current.tone}">{current.verdict}</span>
          <span class="risk-label">Tingkat risiko</span>
        </div>

        <div class="meter" role="img" aria-label={`Tingkat risiko ${current.score} dari 100`}>
          <span class="meter-fill meter-fill--{current.tone}" style:width="{current.score}%"></span>
        </div>

        <ol class="flags">
          {#each current.flags as flag, i (flag.lead)}
            <li>
              <span class="num" aria-hidden="true">{i + 1}</span>
              <span><strong>{flag.lead}</strong> {flag.rest}</span>
            </li>
          {/each}
        </ol>
      </div>
    </div>
  </div>
</div>

<style>
  .scan-wrap { position: relative; padding-top: 18px; }
  .sticker--coral { background: var(--coral); }
  .sticker--yellow { background: var(--yellow); }
  .scan { padding: clamp(1.1rem, 3vw, 1.6rem); box-shadow: var(--shadow-hard-xl); }
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }
  .head h2 { font-size: 1.5em; }
  .pill--soft { background: var(--indigo-tint); color: var(--indigo-text); }

  .tabs { display: flex; flex-wrap: wrap; gap: var(--space-2); margin-bottom: var(--space-4); }
  .tabs button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    flex: 1 1 100px;
    min-height: 52px;
    padding: 0 var(--space-3);
    background: var(--white);
    color: var(--ink);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
    font: inherit;
    font-weight: 700;
    font-size: 0.95em;
    cursor: pointer;
    transition: background-color var(--dur-fast) var(--ease-out);
  }
  .tabs button:hover { background: var(--indigo-tint); }
  .tabs button.is-active { background: var(--ink); color: var(--lime); }
  .tabs button:focus-visible { outline: var(--focus-ring); outline-offset: var(--focus-offset); }

  .sample-label { margin: 0 0 var(--space-2); font-weight: 700; font-size: var(--text-sm); color: var(--muted); }
  .sample {
    margin: 0 0 var(--space-4);
    min-height: 96px;
    padding: var(--space-4) 18px;
    background: var(--paper);
    border: var(--border-width) dashed var(--ink);
    border-radius: var(--radius-md);
    font-size: 1.05em;
    line-height: 1.5;
    overflow-wrap: anywhere;
  }

  .scan-bar {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    min-height: 58px;
    padding: 0 var(--space-4);
    background: var(--ink);
    color: var(--lime);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 1.1em;
    text-decoration: none;
    box-shadow: var(--shadow-hard-xs);
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
  }
  .scan-bar:hover { color: var(--lime); transform: translate(-2px, -2px); box-shadow: var(--shadow-hard-sm); }
  .scan-bar:active { transform: translate(2px, 2px); box-shadow: none; }

  .verdict {
    margin-top: 18px;
    padding: 18px;
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-md);
    background: var(--coral-tint);
  }
  .verdict--warn { background: var(--yellow-tint); }
  .verdict-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }
  .chip {
    display: inline-flex;
    padding: 4px 14px;
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-pill);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 0.95em;
    color: var(--ink);
  }
  .chip--danger { background: var(--coral); }
  .chip--warn { background: var(--yellow); }
  .risk-label { font-weight: 700; font-size: var(--text-sm); }

  .meter {
    height: 20px;
    background: var(--white);
    border: var(--border-width) solid var(--ink);
    border-radius: var(--radius-pill);
    overflow: hidden;
  }
  .meter-fill {
    display: block;
    height: 100%;
    border-right: var(--border-width) solid var(--ink);
    transition: width var(--dur-slow) var(--ease-out);
  }
  .meter-fill--danger { background: var(--coral); }
  .meter-fill--warn { background: var(--yellow); }

  .flags { display: grid; gap: var(--space-3); margin: var(--space-4) 0 0; padding: 0; list-style: none; }
  .flags li { display: flex; align-items: flex-start; gap: var(--space-3); font-size: 0.98em; line-height: 1.4; }
  .num {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    flex: none;
    background: var(--ink);
    color: var(--white);
    border-radius: 50%;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: var(--text-sm);
  }
  strong { font-weight: 700; }
</style>
