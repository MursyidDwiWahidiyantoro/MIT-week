<script lang="ts">
  interface Props {
    label: string;
    id: string;
    value?: string;
    hint?: string;
    error?: string;
    type?: 'text' | 'url';
    placeholder?: string;
    disabled?: boolean;
    oninput?: (value: string) => void;
  }

  let {
    label,
    id,
    value = $bindable(''),
    hint,
    error,
    type = 'text',
    placeholder,
    disabled = false,
    oninput
  }: Props = $props();

  const describedBy = $derived(
    [hint ? `${id}-hint` : '', error ? `${id}-error` : ''].filter(Boolean).join(' ') || undefined
  );
</script>

<div class="field">
  <label for={id}>{label}</label>
  {#if hint}
    <p class="hint" id={`${id}-hint`}>{hint}</p>
  {/if}
  <input
    class="input"
    {id}
    {type}
    {placeholder}
    {disabled}
    bind:value
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={describedBy}
    oninput={(e) => oninput?.(e.currentTarget.value)}
  />
  {#if error}
    <p class="error-text" id={`${id}-error`} role="alert">{error}</p>
  {/if}
</div>

<style>
  .hint { margin: 0; color: var(--muted); font-size: var(--text-sm); }
</style>
