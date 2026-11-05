<script lang="ts">
import { onMount } from 'svelte';
let theme = $state('dark');
function setTheme(value: string) {
 theme = value; document.documentElement.dataset.theme = value;
 document.querySelector('meta[name="theme-color"]')?.setAttribute('content', value === 'dark' ? '#070302' : '#fff8f4');
 try { localStorage.setItem('amba-theme', value); } catch { /* Optional persistence. */ }
}
onMount(() => {try {setTheme(localStorage.getItem('amba-theme') === 'light' ? 'light' : 'dark');} catch {setTheme('dark');}});
</script>
<fieldset class="theme-switch"><legend class="sr-only">Tema warna</legend><button type="button" aria-pressed={theme === 'dark'} onclick={() => setTheme('dark')}>Gelap</button><button type="button" aria-pressed={theme === 'light'} onclick={() => setTheme('light')}>Terang</button></fieldset>
