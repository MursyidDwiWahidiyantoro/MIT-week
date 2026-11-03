#!/usr/bin/env node
/**
 * Fails CI if any text/background pair in tokens.css drops below its required
 * contrast: 4.5:1 for body text, 3:1 for large text, UI borders and focus ring.
 * Reads the real token file — no duplicated colour values.
 */
import fs from 'node:fs';
import path from 'node:path';

const tokenFile = path.resolve('src/lib/styles/tokens.css');
const css = fs.readFileSync(tokenFile, 'utf8');

/** Reads a token and follows `var(--other)` aliases to a literal hex. */
function token(name, seen = new Set()) {
  if (seen.has(name)) throw new Error(`token --${name} is circular`);
  seen.add(name);
  const m = css.match(new RegExp(`--${name}:\\s*([^;]+);`));
  if (!m) throw new Error(`token --${name} not found in ${tokenFile}`);
  const value = m[1].trim();
  const hex = value.match(/^#[0-9a-fA-F]{3,8}$/);
  if (hex) return hex[0];
  const alias = value.match(/^var\(--([a-z0-9-]+)\)$/);
  if (alias) return token(alias[1], seen);
  throw new Error(`token --${name} is neither a hex nor a var() alias: ${value}`);
}

function rgb(hex) {
  let h = hex.replace('#', '');
  if (h.length === 3) h = h.split('').map((c) => c + c).join('');
  return [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16) / 255);
}

function luminance(hex) {
  const [r, g, b] = rgb(hex).map((c) =>
    c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a, b) {
  const [l1, l2] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (l1 + 0.05) / (l2 + 0.05);
}

const surfaces = ['bg', 'surface', 'surface-2'];
const textPairs = [
  ['text', 'bg'], ['text', 'surface'], ['text', 'surface-2'],
  ['text-2', 'bg'], ['text-2', 'surface'], ['text-2', 'surface-2'],
  ['text-3', 'bg'], ['text-3', 'surface'],
  ['accent', 'bg'], ['accent', 'surface'],
  ['accent-2', 'bg'], ['accent-2', 'surface'],
  ['ok', 'bg'], ['warn', 'bg'], ['err', 'bg'],
  ['ok', 'surface'], ['warn', 'surface'], ['err', 'surface'],
  /* sticker palette: ink sits on every saturated fill, white on indigo */
  ['ink', 'paper'], ['ink', 'white'], ['ink', 'lime'], ['ink', 'yellow'],
  ['ink', 'coral'], ['ink', 'indigo-soft'], ['ink', 'indigo-tint'],
  ['ink', 'coral-tint'], ['ink', 'yellow-tint'],
  ['white', 'indigo'], ['on-indigo-text', 'indigo'], ['lime', 'ink'],
  ['indigo-text', 'paper'], ['indigo-text', 'white'], ['indigo-text', 'indigo-tint'],
  ['muted', 'paper'], ['muted', 'white'], ['body-2', 'paper'], ['body-2', 'white']
];

let failures = 0;

for (const [fg, bg] of textPairs) {
  const ratio = contrast(token(fg), token(bg));
  if (ratio < 4.5) {
    console.error(`FAIL  ${fg} on ${bg}: ${ratio.toFixed(2)}:1 (need 4.5)`);
    failures++;
  }
}

for (const bg of surfaces) {
  const ratio = contrast(token('ink'), token(bg));
  if (ratio < 3) {
    console.error(`FAIL  border-strong on ${bg}: ${ratio.toFixed(2)}:1 (need 3.0 for UI borders)`);
    failures++;
  }
}

const focusRatio = contrast(token('ink'), token('bg'));
if (focusRatio < 3) {
  console.error(`FAIL  focus ring on bg: ${focusRatio.toFixed(2)}:1 (need 3.0)`);
  failures++;
}

if (failures) {
  console.error(`\n${failures} contrast failure(s).`);
  process.exit(1);
}
console.log(`contrast OK — ${textPairs.length + surfaces.length + 1} pairs checked`);
