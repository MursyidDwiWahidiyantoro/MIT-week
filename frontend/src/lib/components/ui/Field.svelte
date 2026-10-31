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
    <p class="error" id={`${id}-error`} role="alert">{error}</p>
  {/if}
</div>

<style>
  .field { display: grid; gap: var(--space-2); }
  label { font-weight: 600; font-size: var(--text-sm); }
  .hint { margin: 0; color: var(--text-2); font-size: var(--text-sm); }
  input {
    min-height: 44px;
    padding: 0 var(--space-3);
    background: var(--surface-2);
    color: var(--text);
    border: var(--border-width) solid var(--border-strong);
    border-radius: var(--radius-md);
    font: inherit;
  }
  input:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }
  input[aria-invalid='true'] { border-color: var(--err); }
  .error { margin: 0; color: var(--err); font-size: var(--text-sm); }
</style>
