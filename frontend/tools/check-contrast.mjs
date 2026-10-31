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

function token(name) {
  const m = css.match(new RegExp(`--${name}:\\s*(#[0-9a-fA-F]{3,8})`));
  if (!m) throw new Error(`token --${name} not found in ${tokenFile}`);
  return m[1];
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
  ['ok', 'surface'], ['warn', 'surface'], ['err', 'surface']
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
  const ratio = contrast(token('border-strong'), token(bg));
  if (ratio < 3) {
    console.error(`FAIL  border-strong on ${bg}: ${ratio.toFixed(2)}:1 (need 3.0 for UI borders)`);
    failures++;
  }
}

const focusRatio = contrast(token('accent-2'), token('bg'));
if (focusRatio < 3) {
  console.error(`FAIL  focus ring on bg: ${focusRatio.toFixed(2)}:1 (need 3.0)`);
  failures++;
}

if (failures) {
  console.error(`\n${failures} contrast failure(s).`);
  process.exit(1);
}
console.log(`contrast OK — ${textPairs.length + surfaces.length + 1} pairs checked`);
