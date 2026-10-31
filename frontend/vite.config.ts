import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import { securityHeaders } from './tools/vite-security-headers.ts';

// SvelteKit 3 takes its config through the plugin, not svelte.config.js.
export default defineConfig({
  plugins: [
    sveltekit({
      // SPA: no SSR, single static bundle with a fallback page.
      adapter: adapter({ fallback: 'index.html', strict: false }),
      prerender: { entries: [] },
      // SvelteKit injects an inline bootstrap script; with mode 'hash' it adds
      // that script's sha256 to script-src automatically. A hardcoded
      // script-src 'self' would block hydration and render a blank page.
      csp: {
        mode: 'hash',
        directives: {
          'default-src': ['self'],
          'script-src': ['self'],
          'style-src': ['self', 'unsafe-inline'],
          'img-src': ['self', 'data:', 'blob:'],
          'font-src': ['self'],
          'connect-src': ['self', 'http://localhost:8080'],
          'worker-src': ['self', 'blob:'],
          'frame-ancestors': ['none'],
          'base-uri': ['none'],
          'object-src': ['none']
        }
      }
    }),
    securityHeaders()
  ],
  test: {
    include: ['src/**/*.{test,spec}.{js,ts}'],
    environment: 'jsdom',
    setupFiles: ['./vitest-setup.ts']
  }
});
