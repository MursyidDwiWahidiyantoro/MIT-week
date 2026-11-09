import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import { securityHeaders } from './tools/vite-security-headers.ts';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

/**
 * CSP hash of the inline pre-paint script in src/app.html. SvelteKit only
 * hashes the scripts it injects itself, so this one is hashed here, from the
 * file, and can never drift from its content.
 */
function appHtmlScriptHash(): `sha256-${string}` {
  const html = readFileSync(new URL('./src/app.html', import.meta.url), 'utf8');
  const match = html.match(/<script>([\s\S]*?)<\/script>/);
  if (!match) throw new Error('src/app.html: inline pre-paint script not found');
  return `sha256-${createHash('sha256').update(match[1]).digest('base64')}`;
}

/**
 * The API origin the built app may talk to. CSP connect-src is generated at
 * build time from VITE_API_BASE_URL, so a production build for Vercel pins
 * the Railway/Fly API origin instead of localhost (TECH_SPEC §7.4).
 */
function apiOrigins(): Array<`${string}://${string}.${string}` | `${string}://localhost:${number}`> {
  const raw = process.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1';
  // A relative base ("/api/v1", the Docker setup behind nginx) is same-origin:
  // 'self' already covers it.
  if (raw.startsWith('/')) return [];
  try {
    return [new URL(raw).origin as `${string}://${string}.${string}`];
  } catch {
    throw new Error(`VITE_API_BASE_URL is not a valid URL: ${raw}`);
  }
}

// SvelteKit 3 takes its config through the plugin, not svelte.config.js.
export default defineConfig({
  plugins: [
    sveltekit({
      // SPA: no SSR, single static bundle with a fallback page.
      adapter: adapter({ fallback: 'index.html', strict: false }),
      prerender: { entries: [] },
      alias: { '$lib': 'src/lib' },
      // SvelteKit injects an inline bootstrap script; with mode 'hash' it adds
      // that script's sha256 to script-src automatically. A hardcoded
      // script-src 'self' would block hydration and render a blank page.
      csp: {
        mode: 'hash',
        directives: {
          'default-src': ['self'],
          'script-src': ['self', appHtmlScriptHash()],
          'style-src': ['self', 'unsafe-inline'],
          'img-src': ['self', 'data:', 'blob:'],
          'font-src': ['self'],
          'connect-src': ['self', ...apiOrigins()],
          'worker-src': ['self', 'blob:'],
          // Required by Svelte's runtime and SvelteKit's service-worker
          // registration. Narrowing this list further blocks hydration.
          'trusted-types': ['svelte-trusted-html', 'sveltekit-trusted-url'],
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
  },
  // Component tests mount Svelte in jsdom, so Vitest must load Svelte's
  // browser build; without this `mount` resolves to the server entry and throws.
  resolve: process.env.VITEST ? { conditions: ['browser'] } : undefined
});
