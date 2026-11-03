<script lang="ts">
  /**
   * The one icon set: inline stroke SVG, 24x24 grid, per docs/design-spec.md §6.
   * Decorative by default; pass `label` only when the icon is the sole name of
   * a control.
   */
  export type IconName =
    | 'shield-check'
    | 'chat'
    | 'link'
    | 'phone'
    | 'arrow-right'
    | 'check'
    | 'book'
    | 'speaker'
    | 'moon'
    | 'type'
    | 'sparkle'
    | 'image'
    | 'text'
    | 'alert'
    | 'lock'
    | 'download'
    | 'copy'
    | 'print'
    | 'refresh'
    | 'whatsapp'
    | 'volume-off'
    | 'chevron';

  interface Props {
    name: IconName;
    size?: number;
    label?: string;
  }

  let { name, size = 24, label }: Props = $props();

  const filled = $derived(name === 'sparkle');
  const strokeWidth = $derived(name === 'shield-check' ? 2.4 : 2.2);
</script>

<svg
  viewBox="0 0 24 24"
  width={size}
  height={size}
  fill={filled ? 'currentColor' : 'none'}
  stroke={filled ? 'none' : 'currentColor'}
  stroke-width={strokeWidth}
  stroke-linecap="round"
  stroke-linejoin="round"
  role={label ? 'img' : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : 'true'}
  focusable="false"
>
  {#if name === 'shield-check'}
    <path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6l7-3z" />
    <path d="M9 12l2 2 4-4" />
  {:else if name === 'chat'}
    <path d="M4 5h16v11H9l-5 4V5z" />
  {:else if name === 'link'}
    <path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1" />
    <path d="M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1" />
  {:else if name === 'phone'}
    <path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A15 15 0 0 1 3 6a2 2 0 0 1 2-2z" />
  {:else if name === 'arrow-right'}
    <path d="M5 12h14M13 6l6 6-6 6" />
  {:else if name === 'check'}
    <path d="M5 12l5 5 9-10" />
  {:else if name === 'book'}
    <path d="M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2V5z" />
    <path d="M8 7h7" />
  {:else if name === 'speaker'}
    <path d="M4 10v4h4l5 4V6L8 10H4z" />
    <path d="M16 9a4 4 0 0 1 0 6" />
  {:else if name === 'volume-off'}
    <path d="M4 10v4h4l5 4V6L8 10H4z" />
    <path d="M16 10l4 4M20 10l-4 4" />
  {:else if name === 'moon'}
    <path d="M20 14.5A8 8 0 0 1 9.5 4 8 8 0 1 0 20 14.5z" />
  {:else if name === 'type'}
    <path d="M5 19L11 5l6 14M7.5 14h7" />
  {:else if name === 'sparkle'}
    <path d="M12 2l2.6 7.4L22 12l-7.4 2.6L12 22l-2.6-7.4L2 12l7.4-2.6z" />
  {:else if name === 'image'}
    <rect x="3.5" y="4.5" width="17" height="15" rx="2" />
    <circle cx="9" cy="10" r="1.6" />
    <path d="M4 17l5-5 4 4 3-2 4 3" />
  {:else if name === 'text'}
    <path d="M4 6h16M4 12h16M4 18h10" />
  {:else if name === 'alert'}
    <path d="M12 4l9 16H3l9-16z" />
    <path d="M12 10v4M12 17.5v.5" />
  {:else if name === 'lock'}
    <rect x="4.5" y="10.5" width="15" height="10" rx="2" />
    <path d="M8 10.5V8a4 4 0 0 1 8 0v2.5" />
  {:else if name === 'download'}
    <path d="M12 4v11M7 11l5 5 5-5M5 20h14" />
  {:else if name === 'copy'}
    <rect x="8" y="8" width="12" height="12" rx="2" />
    <path d="M16 8V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3" />
  {:else if name === 'print'}
    <path d="M7 9V4h10v5" />
    <rect x="4" y="9" width="16" height="7" rx="1.5" />
    <path d="M7 16h10v4H7z" />
  {:else if name === 'refresh'}
    <path d="M4 12a8 8 0 0 1 13.7-5.6L20 8" />
    <path d="M20 4v4h-4" />
    <path d="M20 12a8 8 0 0 1-13.7 5.6L4 16" />
    <path d="M4 20v-4h4" />
  {:else if name === 'whatsapp'}
    <path d="M20 11.5a8 8 0 0 1-11.8 7L4 20l1.5-4.1A8 8 0 1 1 20 11.5z" />
    <path d="M9 9.5c0 3 2.5 5.5 5.5 5.5" />
  {:else if name === 'chevron'}
    <path d="M6 9l6 6 6-6" />
  {/if}
</svg>
