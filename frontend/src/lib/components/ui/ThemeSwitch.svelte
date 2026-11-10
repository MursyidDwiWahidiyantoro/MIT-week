<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { onMount } from 'svelte';

  /**
   * Theme: the device decides unless the visitor picks one (Sir, 10 Nov 2026).
   * 'system' follows prefers-color-scheme live; the pre-paint script in
   * app.html applies the same rule before the first paint.
   */
  type Pref = 'system' | 'light' | 'dark';
  const KEY = 'amba-theme';
  let pref = $state<Pref>('system');
  let media: MediaQueryList | undefined;

  function apply(next: Pref) {
    const dark = next === 'dark' || (next === 'system' && !!media?.matches);
    const root = document.documentElement;
    root.dataset.theme = dark ? 'dark' : 'light';
    root.dataset.themePref = next;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', dark ? '#1e1d1b' : '#f4f3ee');
  }

  function choose(next: Pref) {
    pref = next;
    apply(next);
    try {
      if (next === 'system') localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, next);
    } catch {
      // Session-only when storage is blocked.
    }
  }

  onMount(() => {
    media = window.matchMedia('(prefers-color-scheme: dark)');
    let stored: string | null;
    try {
      stored = localStorage.getItem(KEY);
    } catch {
      stored = null;
    }
    pref = stored === 'light' || stored === 'dark' ? stored : 'system';
    apply(pref);
    const follow = () => {
      if (pref === 'system') apply('system');
    };
    media.addEventListener('change', follow);
    return () => media?.removeEventListener('change', follow);
  });

  const OPTIONS: Array<{ value: Pref; label: 'theme.system' | 'theme.light' | 'theme.dark' }> = [
    { value: 'system', label: 'theme.system' },
    { value: 'light', label: 'theme.light' },
    { value: 'dark', label: 'theme.dark' }
  ];
</script>

<fieldset class="theme-switch">
  <legend class="sr-only">{t('theme.legend')}</legend>
  {#each OPTIONS as option (option.value)}
    <button type="button" aria-pressed={pref === option.value} onclick={() => choose(option.value)}>
      {t(option.label)}
    </button>
  {/each}
</fieldset>
