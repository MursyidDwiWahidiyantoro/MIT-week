<script lang="ts">
  import { t } from '$lib/i18n/index.svelte';
  import { page } from '$app/state';
  import { resolve } from '$app/paths';
</script>

<svelte:head><title>{t('error.title404')}</title></svelte:head>

<section class="state">
  <p class="eyebrow">{page.status === 404 ? '404' : t('error.eyebrow')}</p>
  <h1 class="h1">{page.status === 404 ? t('error.h404') : t('error.hOther')}</h1>
  <p class="lede">
    {#if page.status === 404}
      {t('error.lede404')}
    {:else}
      {t('error.ledeOther')}
    {/if}
  </p>

  <div class="actions">
    <a class="btn btn--primary" href={resolve('/')}>{t('error.home')}</a>
    <a class="btn btn--quiet" href={resolve('/deteksi')}>{t('error.check')}</a>
  </div>
</section>

<style>
  .state { max-width: 44rem; padding-block: var(--space-7); }
  h1 { margin-bottom: var(--space-4); }
  .lede { margin-bottom: var(--space-6); }
  .actions { display: flex; gap: var(--space-3); flex-wrap: wrap; }
</style>
