<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { onMount } from 'svelte';
  import Icon from './Icon.svelte';

  /**
   * Theme: dark by default, light on request (Sir, 10 Nov 2026). One toggle,
   * as Gemini puts it in its menu. The pre-paint script in app.html applies
   * the stored choice before the first paint.
   */
  interface Props {
    /** Show the text label (the side rail); icon only otherwise. */
    withLabel?: boolean;
  }
  let { withLabel = false }: Props = $props();

  type Theme = 'dark' | 'light';
  const KEY = 'amba-theme';
  let theme = $state<Theme>('dark');

  function choose(next: Theme) {
    theme = next;
    document.documentElement.dataset.theme = next;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', next === 'dark' ? '#0f0f0f' : '#faf9f9');
    try {
      if (next === 'dark') localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, next);
    } catch {
      // Session-only when storage is blocked.
    }
  }

  onMount(() => {
    theme = document.documentElement.dataset.theme === 'light' ? 'light' : 'dark';
  });

  const label = $derived(theme === 'dark' ? t('theme.toLight') : t('theme.toDark'));
</script>

<button
  type="button"
  class={withLabel ? 'nav-link theme-toggle' : 'icon-btn theme-toggle'}
  aria-label={withLabel ? undefined : label}
  title={label}
  onclick={() => choose(theme === 'dark' ? 'light' : 'dark')}
>
  <Icon name={theme === 'dark' ? 'sun' : 'moon'} size={20} />
  {#if withLabel}<span class="lbl">{label}</span>{/if}
</button>

<style>
  .theme-toggle.nav-link {
    width: 100%;
    border: 0;
    background: transparent;
    font: inherit;
    font-size: var(--text-sm);
    font-weight: 500;
    cursor: pointer;
    text-align: left;
  }
</style>
