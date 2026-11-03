# AMBA — Visual Design System (implementation record)

> Status: **implemented** in `frontend/`. Source of truth for the look:
> `docs/waspada-design-spec.md`. This file records how that spec maps onto the
> real SvelteKit app, and the decisions taken where the spec left room.

---

## 1. Concept in one line

A **neo-brutalist sticker sheet**: every surface is a flat saturated sticker with
a thick ink outline and a hard, blur-free shadow. It lifts when you point at it
and squishes when you press it. Bold enough for a feed, plain enough for a
grandparent.

**Motion verb: press.** Nothing slides or fades decoratively. Things are peeled
up (hover), pressed flat (active), or bob in place (stickers).

**Signature moment:** the scanner card in the hero. Three tabs, each a real
scam pattern, with a risk meter that fills and a verdict panel that retints. It
teaches the shape of a verdict before the visitor submits anything.

---

## 2. Tokens

All tokens live in `frontend/src/lib/styles/tokens.css`. Components reference
names only; no raw hex outside that file.

### Colour

| Token | Value | Role |
|---|---|---|
| `--ink` | `#12142B` | text, every border, every hard shadow |
| `--paper` | `#F7F7FB` | page ground |
| `--white` | `#FFFFFF` | cards, inputs, quiet buttons |
| `--indigo` | `#4A44EE` | primary action, CTA panel |
| `--indigo-soft` | `#DAD8FF` | decorative field, big-text tile |
| `--indigo-tint` | `#EDEBFF` | selected tab, example pill |
| `--indigo-text` | `#3A34C4` | eyebrow labels |
| `--lime` | `#D4F54A` | highlight, positive, ink-panel accent |
| `--yellow` | `#FFC93C` | suspicious |
| `--coral` | `#FF5C4D` | danger |
| `--coral-tint` | `#FFE3DF` | danger verdict panel |
| `--yellow-tint` | `#FFF3CF` | suspicious verdict panel |
| `--muted` | `#4A4E69` | secondary text |
| `--body-2` | `#2B2E4A` | paragraph text |
| `--ok` / `--warn` / `--err` | `#1F6B33` / `#7A5200` / `#B3261E` | status **text** only |

**Two deliberate departures from the source spec, both for contrast:**

1. The spec's saturated coral/yellow/lime are used as **fills** only, with ink on
   top. Status *text* uses darkened variants so it passes AA on paper and white.
   Ink on coral is 5.94:1, ink on yellow 11.78:1, ink on lime 14.61:1.
2. `--indigo` is never used for body text (5.86:1 on paper); `--indigo-text` is
   the text-safe indigo at 7.96:1.

`tools/check-contrast.mjs` reads these tokens (following `var()` aliases) and
fails CI on 41 pairs, including the whole sticker palette.

### Type

| Role | Family | Notes |
|---|---|---|
| Display, headings, buttons, labels | **Bricolage Grotesque** (variable) | weight 800, tracking −0.025em to −0.03em |
| Body, UI | **Atkinson Hyperlegible** | 400/700, chosen for low-vision legibility |

Two families, both self-hosted via `@fontsource` so the CSP stays `font-src 'self'`.

Scale: `--text-display` `clamp(2.5rem, 6vw, 4.6rem)` for H1, `--text-2xl`
`clamp(1.75rem, 4.4vw, 3.2rem)` for H2, body 18px/1.5, small 15px.

**No monospace anywhere.** The spec's "technical" labels use the display face at
small size instead: a monospace label is the single most common generated-page
tell, and it carried no information here.

### Edges and shadows

- Border `2.5px solid var(--ink)` on every card, button, pill, tab, input.
- Radii: cards 26px, buttons 16–18px, pills 99px, chat bubble `22px 22px 22px 6px`.
- Shadows are **always** `Npx Npx 0 var(--ink)`: 3px chips, 6px buttons, 8px
  cards, 10px hero card and CTA panel. Dark/accent panels use an indigo shadow.
- Blur is zero everywhere. There is no `box-shadow` with a blur radius in the
  codebase, except the proof-card exporter's print surface.

---

## 3. Interaction language

One treatment per role, reused everywhere:

| Role | Rest | Hover | Active |
|---|---|---|---|
| Card (`.lift`) | hard shadow | `translate(-3px,-3px)` + 9px shadow | `translate(3px,3px)` + no shadow |
| Button (`.btn`) | hard shadow | same peel, shadow grows | squish flat |
| Nav link (`.nav-link`) | transparent | lime fill + ink border + peel | 1px press |
| Tab | white | indigo tint | ink fill, lime text |
| Input | dashed ink border | — | solid border on focus |

- Focus: `4px solid var(--ink)` with 3px offset, on every interactive element.
- Press feedback is never removed, in any mode.
- **Idle motion:** the two hero stickers bob on a 3.6s loop and the marquee
  runs continuously. Both stop under reduced motion and in Esensial mode.

---

## 4. Motion budget

| Token | Value |
|---|---|
| `--dur-fast` | 120ms |
| `--dur-base` | 200ms |
| `--dur-slow` | 420ms |
| `--ease-out` | `cubic-bezier(.22,1,.36,1)` |

Only `transform`, `opacity` and `background-color` animate. No layout property
is animated anywhere.

**Calm is the default.** `<html data-motion="calm">` ships in `app.html`, so the
first paint is already still; `mode.svelte.ts` only upgrades it to `full` when
the visitor is in Jelajah *and* has not asked for reduced motion. A visitor with
JS blocked, or a crawler, gets the calm version: all content visible, zero
animation.

---

## 5. Layout

- Max width 1240px, 20px gutter at every size.
- Fluid `flex-wrap` columns with `flex: 1 1 <basis>` and `min-width: 0`; no
  breakpoint-heavy grids.
- Every section uses a different composition: split hero, marquee band, 3-card
  row, 4-card bento, copy-left/tiles-right, full-bleed CTA panel.
- Phone header is exactly three rows (brand / nav / tools) and the nav stays on
  **one** line down to 360px.

---

## 6. Accessibility contract

- Every route: one `<h1>`, ordered headings, landmarks, `lang="id"`.
- Skip link is the first tab stop on every page.
- Tap targets: measured minimum 44×44px; the audit reports zero below that.
- Status is never colour alone: every verdict pairs an icon, a text label and a
  written explanation.
- Tabs use `role="tablist"` / `role="tab"` / `aria-selected` with roving
  tabindex and arrow keys.
- The first-visit mode picker is **inline, not modal**: it sits after the hero,
  does not trap Tab and does not steal focus on load.
- Esensial mode: single column, +12.5% type, all motion off, identical content.

---

## 7. Honest limits

- The hero scanner is a **worked example**, labelled "Contoh", not a live scan.
  It makes no network call.
- Demo dossiers carry a visible "CONTOH / SIMULASI" banner.
- The score is a risk score, never a confidence score, and never rendered with
  a percent sign.
- "Mode tenang" and "Bacakan hasil" are real (motion off, `speechSynthesis`).
  "Kumpulkan lencananya" is illustrative only; there is no badge persistence,
  by design (no accounts, no storage).
