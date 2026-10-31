#!/usr/bin/env node
/**
 * Fails if the initial JS payload exceeds 250 KB gzip (CLAUDE.md, TECH_SPEC §8).
 * Measures the built entry chunks; assets are read from the adapter-static output.
 */
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';

const LIMIT = 250 * 1024;
const dirs = ['build/_app/immutable', '.svelte-kit/output/client/_app/immutable'];

let total = 0;
let found = false;

for (const dir of dirs) {
  if (!fs.existsSync(dir)) continue;
  found = true;
  const walk = (d) => {
    for (const entry of fs.readdirSync(d, { withFileTypes: true })) {
      const p = path.join(d, entry.name);
      if (entry.isDirectory()) walk(p);
      else if (entry.name.endsWith('.js')) total += zlib.gzipSync(fs.readFileSync(p)).length;
    }
  };
  walk(dir);
  break;
}

if (!found) {
  console.error('No build output found. Run `npm run build` first.');
  process.exit(1);
}

const kb = (total / 1024).toFixed(1);
if (total > LIMIT) {
  console.error(`FAIL  initial JS ${kb} KB gz exceeds ${LIMIT / 1024} KB`);
  process.exit(1);
}
console.log(`size OK — ${kb} KB gz (limit ${LIMIT / 1024} KB)`);
