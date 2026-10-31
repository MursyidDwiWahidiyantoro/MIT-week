import type { Config } from 'tailwindcss';

/**
 * Every colour maps to a CSS custom property defined in tokens.css.
 * Components must use these names, never raw hex values.
 */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--bg)',
        surface: 'var(--surface)',
        'surface-2': 'var(--surface-2)',
        border: 'var(--border)',
        'border-strong': 'var(--border-strong)',
        text: 'var(--text)',
        'text-2': 'var(--text-2)',
        'text-3': 'var(--text-3)',
        accent: 'var(--accent)',
        'accent-2': 'var(--accent-2)',
        ok: 'var(--ok)',
        warn: 'var(--warn)',
        err: 'var(--err)'
      },
      fontFamily: {
        sans: 'var(--font-sans)',
        mono: 'var(--font-mono)'
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)'
      },
      transitionDuration: {
        fast: 'var(--dur-fast)',
        base: 'var(--dur-base)'
      }
    }
  },
  plugins: []
} satisfies Config;
