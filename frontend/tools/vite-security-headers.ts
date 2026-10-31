import type { Plugin, PreviewServer, ViteDevServer } from 'vite';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Applies frontend/security-headers.json to dev and preview responses.
 * The JSON file is the single source of truth; host-specific emission is
 * deferred until deployment is in scope.
 */
export function securityHeaders(): Plugin {
  const file = path.resolve(process.cwd(), 'security-headers.json');

  function apply(server: ViteDevServer | PreviewServer) {
    server.middlewares.use((_req, res, next) => {
      try {
        const headers = JSON.parse(fs.readFileSync(file, 'utf8')) as Record<string, string>;
        for (const [key, value] of Object.entries(headers)) res.setHeader(key, value);
      } catch {
        // A missing or invalid file must not break dev; CI checks it separately.
      }
      next();
    });
  }

  return {
    name: 'amba-security-headers',
    configureServer: apply,
    configurePreviewServer: apply
  };
}
