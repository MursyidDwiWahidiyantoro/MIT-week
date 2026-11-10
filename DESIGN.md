---
name: AMBA
description: Cek dulu sebelum percaya. A content checker for Indonesian families, built from frosted glass over warm neutrals.
colors:
  terracotta: "#a9502e"
  terracotta-hover: "#94452a"
  terracotta-deep: "#7e3b24"
  terracotta-night: "#d97757"
  terracotta-text: "#a1462a"
  ivory-ground: "#f4f3ee"
  ivory-surface: "#fbfaf7"
  ivory-inset: "#efede6"
  charcoal-ink: "#1f1e1c"
  charcoal-body: "#45443f"
  charcoal-muted: "#5d5c56"
  hairline: "#dad7ce"
  hairline-strong: "#8f8c82"
  night-ground: "#1e1d1b"
  night-surface: "#2a2926"
  night-ink: "#f2f0e9"
  verdict-ok: "#2b6646"
  verdict-warn: "#8a5300"
  verdict-err: "#a3302a"
typography:
  display:
    fontFamily: "Bricolage Grotesque Variable, Helvetica Neue, Arial, sans-serif"
    fontSize: "clamp(2.4rem, 1.8rem + 2.7vw, 4.75rem)"
    fontWeight: 600
    lineHeight: 1.04
    letterSpacing: "-0.035em"
  headline:
    fontFamily: "Bricolage Grotesque Variable, Helvetica Neue, Arial, sans-serif"
    fontSize: "clamp(2rem, 1.6rem + 1.7vw, 3.35rem)"
    fontWeight: 600
    lineHeight: 1.04
    letterSpacing: "-0.022em"
  title:
    fontFamily: "Bricolage Grotesque Variable, Helvetica Neue, Arial, sans-serif"
    fontSize: "clamp(1.6rem, 1.25rem + 1.2vw, 2.6rem)"
    fontWeight: 600
    lineHeight: 1.12
    letterSpacing: "-0.022em"
  body:
    fontFamily: "Atkinson Hyperlegible, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 400
    lineHeight: 1.55
  label:
    fontFamily: "Atkinson Hyperlegible, system-ui, -apple-system, Segoe UI, sans-serif"
    fontSize: "0.9375rem"
    fontWeight: 700
    lineHeight: 1.2
rounded:
  sm: "10px"
  md: "14px"
  lg: "22px"
  xl: "28px"
  pill: "999px"
spacing:
  1: "0.25rem"
  2: "0.5rem"
  3: "0.75rem"
  4: "1rem"
  5: "1.5rem"
  6: "2rem"
  7: "3rem"
  8: "4rem"
components:
  button-primary:
    backgroundColor: "{colors.terracotta}"
    textColor: "{colors.ivory-surface}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: "52px"
    padding: "0 1.5rem"
  button-primary-hover:
    backgroundColor: "{colors.terracotta-hover}"
  button-primary-active:
    backgroundColor: "{colors.terracotta-deep}"
  button-glass:
    textColor: "{colors.charcoal-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.md}"
    height: "52px"
    padding: "0 1.5rem"
  input-field:
    backgroundColor: "{colors.ivory-surface}"
    textColor: "{colors.charcoal-ink}"
    rounded: "{rounded.md}"
    height: "54px"
    padding: "12px 1rem"
  nav-link:
    textColor: "{colors.charcoal-body}"
    typography: "{typography.label}"
    rounded: "{rounded.sm}"
    height: "44px"
    padding: "0 14px"
  chip:
    textColor: "{colors.charcoal-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    height: "32px"
    padding: "3px 12px"
---

# Design System: AMBA

## Overview

**Creative North Star: "The Inspection Glass"**

What a family wants to verify sits behind frosted glass, and AMBA's verdict is what clears it. Glass is the instrument in this system, not decoration. It appears where real content passes under a persistent surface or where something is being examined:
- the sticky header;
- the checker;
- the verdict lens over the user's own image or video frame;
- the error pane.

Every other surface is a quieter translucent pane, or plain type on the page's warm light.

The palette is Claude-like and restrained, as pinned by Sir on 10 Nov 2026: ivory and charcoal neutrals and one terracotta accent. Light and dark are both glass, and the device chooses between them. The density is that of a calm tool for older readers: generous type, short measures, one primary action per screen. The rejected look is the earlier one: a black page, neon orange, an accented word in the headline, three numbered step cards and a closing CTA band.

The qualitative language here (the North Star and the names) was derived from Sir's pinned brief and the direction contract; it was not confirmed in a separate interview.

**Key Characteristics:**
- Frosted glass is reserved for surfaces with something under them.
- One terracotta accent, used on the primary action, focus and the risk meter.
- Verdict colours (ok, warn, err) always come with an icon and words.
- Lists sit on hairlines instead of grids of equal cards.
- The page light is three soft colour fields and a fine grain, fixed to the viewport.

## Colors

Warm neutrals with a single terracotta voice; the verdict states are the only other colours.

### Primary
- **Terracotta** (#a9502e, dark theme #d97757): the primary button, the brand tile, focus rings, links, the risk meter and error-countdown fills. In the light theme it is deepened so white text clears 4.5:1 (#a9502e on ivory is 4.88:1).

### Neutral
- **Ivory Ground** (#f4f3ee): the page in the light theme.
- **Ivory Surface** (#fbfaf7): what a glass pane composites to; it is used for contrast checks and solid fallbacks.
- **Charcoal Ink** (#1f1e1c): text in the light theme.
- **Charcoal Body** (#45443f) and **Charcoal Muted** (#5d5c56): secondary and tertiary text, both at least 4.5:1 on the ground.
- **Hairline** (#dad7ce): list rules and field borders.
- **Night Ground** (#1e1d1b), **Night Surface** (#2a2926), **Night Ink** (#f2f0e9): the dark theme, warm charcoal rather than black.

### Verdict roles
- **Ok** (#2b6646), **Warn** (#8a5300), **Err** (#a3302a): the text colours of the three conclusions. Each has a pale fill (`--safe`, `--yellow`, `--coral`) that turns into a dark tint in the dark theme.

### Named Rules
**The One Voice Rule.** Terracotta marks the action and the focus, nothing else. A selected tab, a language toggle or a nav item is a raised neutral segment, never an orange fill.

**The Composite Rule.** A colour token that the contrast check reads is the opaque colour a glass pane composites to. Translucency lives only in the `--glass-*` tokens.

## Typography

**Display Font:** Bricolage Grotesque Variable (with Helvetica Neue, Arial)
**Body Font:** Atkinson Hyperlegible (with system-ui)

**Character:** A grotesque display face with a little warmth in its curves, set at 600 instead of a heavy black. It sits beside a reading face designed for low-vision legibility, chosen for older family members.

### Hierarchy
- **Display** (600, clamp 2.4rem to 4.75rem, line height 1.04, tracking -0.035em): the home headline only, at most two lines.
- **Headline** (600, clamp 2rem to 3.35rem): page titles.
- **Title** (600, clamp 1.6rem to 2.6rem): section headings, and the verdict sentence (at 650).
- **Body** (400, 1.125rem, line height 1.55): reading text, measure 36 to 44ch on the hero and 36em elsewhere.
- **Label** (700, 0.9375rem): buttons, nav, field labels, list item names. Labels are never uppercase or tracked.

### Named Rules
**The No-Eyebrow Rule.** A heading carries its own weight. Nothing sits above it as a kicker; a status code or category goes below the copy.

## Layout

The container is 1280px wide at most, with a fluid gutter from 20px to 64px. The home hero splits 7/12 and 5/12 on desktop and stacks on phones below 60rem. Section rhythm varies: the hero is airy, the checks list is denser, and the learning link is one quiet row. Lists of equal items use a two- or three-column hairline layout; every item starts with a 1px rule, and the column gap is 1.5rem to 3rem. The phone header is three rows: the brand and display toggle, the three destinations, then the language and theme controls on demand.

## Elevation & Depth

Depth is light, not lift.
- Glass panes float on a soft two-part shadow lit from above, with a 1px inner highlight.
- Translucent panes rest almost flat.
- Hover changes light (background or shadow), never position, apart from a 1px press.

### Shadow Vocabulary
- **Glass** (`--glass-shadow`: `0 1px 2px rgba(31,30,28,.06), 0 22px 44px -26px rgba(31,30,28,.32)`; deeper in dark): frosted surfaces only.
- **Raised** (`--shadow-2`: `0 1px 2px rgba(31,30,28,.06), 0 10px 24px -14px rgba(31,30,28,.22)`): the primary button and hovered controls.
- **Resting** (`--shadow-1`: `0 1px 2px rgba(31,30,28,.06)`): translucent panes and selected segments.

### Named Rules
**The Earned Blur Rule.** `backdrop-filter` appears only on the header, the checker, the verdict lens and the error pane. A pane that has nothing under it does not blur.

## Shapes

Radius is small for controls and grows with containment:
- 10px: nav items and small buttons;
- 14px: buttons, fields and segment tracks;
- 22px: panes;
- 28px: the verdict lens.

Pills are kept for chips and the jump links only. Borders are 1px hairlines; no element has a thick coloured side stripe.

## Components

### Buttons
- **Primary:** terracotta fill, white text in the light theme and charcoal in the dark theme, 14px radius, 52px tall. Hover deepens the fill; press moves down 1px.
- **Glass (default):** a strong glass fill, a hairline edge, ink text. Hover brightens it to the surface colour.
- **Quiet:** transparent with a hairline, for secondary actions such as Tetap di sini.

### Chips
- Status and evidence chips are neutral when they mark completion. A chip takes a verdict colour only when it states a verdict, and then always with an icon.

### Cards / Containers
- **Panel:** a translucent pane without blur, with a hairline edge and an inner highlight, for result sections.
- **Glass:** the frosted pane, for the four earned surfaces.

### Inputs / Fields
- The fill is solid (`--field-bg`) for legibility, with a hairline border that strengthens on hover. Focus is a terracotta border plus a 3px soft ring. Placeholders use the muted colour at full opacity (5.9:1).

### Navigation
- The logo, the wordmark and the three destinations are grouped on the left of a sticky glass bar, and the display controls sit on the right. The current page is a raised neutral segment.

### The Verdict Lens (signature)
- The result's verdict sits on a strong glass pane over the user's own content: the uploaded image or the video's first frame, slightly blurred and scaled at the frame.
- The frost starts at 30px and clears to 20px as the verdict lands, while the picture settles from 1.08 to 1. This takes 520ms with ease-out and is instant under reduced motion.
- Without a picture, the lens sits over a field of the verdict's own colour. The score number and the meter take the verdict colour.

### Error Pane
- A centred glass pane: a plain sentence, the status code under the copy, and a six-second countdown home with a linearly depleting bar. "Ke beranda sekarang" leaves at once and "Tetap di sini" stops the timer (WCAG 2.2.1).

## Do's and Don'ts

### Do:
- **Do** show the thing being checked beside its verdict.
- **Do** keep terracotta for the primary action and focus.
- **Do** put lists of equal items on hairlines.
- **Do** theme browser surfaces from the palette: selection, caret, scrollbar, focus.
- **Do** check both themes. The device picks the theme by default, and the visitor can choose Otomatis, Terang or Gelap.

### Don't:
- **Don't** put a kicker or eyebrow label above a heading.
- **Don't** blur a pane that has nothing under it, or turn every surface into glass.
- **Don't** use a black page with a neon accent, an accented word in the headline, three numbered step cards, or a closing CTA band.
- **Don't** use a coloured side stripe, gradient text, glow as depth, or hard offset shadows.
- **Don't** animate width, height or padding. Use transform and opacity.
