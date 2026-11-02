# Design Specification for Implementation

> Audience: Claude Code (or any coding agent) implementing this design in a **plain project** (vanilla HTML + CSS + JS, no framework required).
> Product: an **AI-powered scam scanner + scam-education website**. This spec is **brand-neutral**: it only describes the design. Everywhere you see `[Brand Name]`, use the project's own name and logo. Do not invent or add any brand.
> Source of truth: a finished visual preview. This document describes it precisely so you can rebuild it without seeing it.
> Scope of the preview: **visual only**. Scanning is a mock. Implement the visuals and the small UI interactions listed in section 9. Do not build a real scanning backend unless asked.

---

## 1. Design Concept in One Paragraph

A **neo-brutalist "sticker" style**: thick dark outlines, flat saturated colors, hard offset shadows (no blur), rounded corners, playful tilted stickers, and big friendly headlines. It feels fresh and bold for Gen-Z, while staying **clear and safe for older users** through large text, very high contrast, generous tap targets, plain language, and a built-in text-size toggle. Every card looks like a physical sticker you can press: it lifts on hover and "squishes" when pressed.

**Style keywords:** chunky, friendly, high-contrast, flat color, hard shadows, rounded, playful, readable, trustworthy.

**Avoid:** gradient washes, glassmorphism, soft blurry shadows, thin light-gray text, tiny tap targets, emoji as icons, stock-photo people, jargon.

---

## 2. Design Tokens

Define these as CSS custom properties on `:root` and use them everywhere.

### 2.1 Colors

| Token | Hex | Use |
|---|---|---|
| `--ink` | `#12142B` | Text, all borders, hard shadows, dark panels |
| `--paper` | `#F7F7FB` | Page background |
| `--white` | `#FFFFFF` | Cards, inputs, secondary buttons |
| `--indigo` | `#4A44EE` | Primary brand color, primary button, number badge, CTA panel |
| `--indigo-soft` | `#DAD8FF` | Decorative circle, "Big-text mode" tile |
| `--indigo-tint` | `#EDEBFF` | Small "Live example" pill background |
| `--indigo-text` | `#3A34C4` | Eyebrow labels (small uppercase section labels) |
| `--lime` | `#D4F54A` | Highlight, "safe/positive" accent, dark-panel accent text, button on indigo |
| `--yellow` | `#FFC93C` | Warning / suspicious, sticker, number badge |
| `--coral` | `#FF5C4D` | Danger / scam, sticker |
| `--coral-tint` | `#FFE3DF` | Verdict panel background when danger |
| `--yellow-tint` | `#FFF3CF` | Verdict panel background when suspicious |
| `--muted` | `#4A4E69` | Secondary small text and labels |
| `--body-2` | `#2B2E4A` | Paragraph text (slightly softer than ink) |
| `--ink-soft` | `#3A3E63` | Track of progress bars on dark panel |
| `--on-indigo-text` | `#F0EFFF` | Paragraph text on indigo panel |

**Contrast rules (all verified to pass WCAG AA):** ink on paper/white/lime/yellow/coral; white on indigo `#4A44EE` (about 6.3:1); muted on paper (about 7.7:1). Do not put white text on lime/yellow/coral. Do not lighten muted text.

### 2.2 Typography

Load from Google Fonts:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Atkinson+Hyperlegible:wght@400;700&family=Bricolage+Grotesque:wght@600;700;800&display=swap">
```

- **Display / headings / buttons:** `'Bricolage Grotesque', sans-serif`, weight **800** (700 for marquee), letter-spacing **-0.02em to -0.03em**, tight line-height (1.02 to 1.1).
- **Body / UI text:** `'Atkinson Hyperlegible', system-ui, sans-serif`, weights 400 and 700. (Atkinson is designed for low-vision readability, which is the point.)
- **Base line-height:** 1.5.

**Type scale (all in `em`, relative to the root wrapper's font-size, which the text-size toggle changes):**

| Element | Size |
|---|---|
| Hero H1 | `clamp(2.5em, 6vw, 4.6em)`, weight 800, line-height 1.02 |
| Section H2 | `clamp(2em, 4.4vw, 3.2em)`, line-height 1.05 |
| CTA H2 | `clamp(2em, 4.6vw, 3.4em)` |
| Card H3 | `1.3em` to `1.6em` |
| Hero paragraph | `1.25em` |
| Section paragraph | `1.15em` |
| Body | `1em` |
| Small text / labels | `0.8em` to `0.95em` |
| Eyebrow label | `0.9em`, uppercase, `letter-spacing: 0.08em`, bold, color `--indigo-text` |
| Marquee | `1.5em`, weight 700 |

**Text-size toggle base sizes (px, set on the page wrapper):** `A = 18px`, `A+ = 20.5px`, `A++ = 23px`. Because everything is in `em`, the whole page scales. **Important:** set `font-size` only on leaf text elements, not on intermediate containers, to avoid `em` compounding.

### 2.3 Borders, Radii, Shadows

- **Border:** `2.5px solid var(--ink)` on every card, button, pill, tab, badge, input box.
- **Radii:**
  - Large cards: `26px` to `28px`; CTA panel `36px`
  - Number badges / icon tiles: `14px` to `20px`
  - Buttons: `16px` to `18px`
  - Pills / chips / badges: `99px` (fully round)
  - Chat bubbles: `22px 22px 22px 6px` (sharp bottom-left tail)
- **Hard offset shadows (no blur):** `box-shadow: Xpx Xpx 0 var(--ink)`
  - Small chips/stickers: `3px 3px 0`
  - Buttons: `5px 5px 0` or `6px 6px 0`
  - Cards: `6px 6px 0` / `8px 8px 0`
  - Hero scanner card: `10px 10px 0`; CTA panel: `10px 10px 0`
  - Some dark/accent panels use an indigo shadow (`#4A44EE`) instead of ink.
- **Spacing rhythm:** gaps of `12 / 16 / 20 / 28 / 40 / 48 / 56px`; card padding `24px` to `30px`; section vertical padding about `88px` to `96px`.
- **Max content width:** `1240px`, centered, side padding `20px`.

### 2.4 Suggested CSS skeleton

```css
:root{
  --ink:#12142B; --paper:#F7F7FB; --white:#fff;
  --indigo:#4A44EE; --indigo-soft:#DAD8FF; --indigo-tint:#EDEBFF; --indigo-text:#3A34C4;
  --lime:#D4F54A; --yellow:#FFC93C; --coral:#FF5C4D;
  --coral-tint:#FFE3DF; --yellow-tint:#FFF3CF;
  --muted:#4A4E69; --body-2:#2B2E4A; --ink-soft:#3A3E63;
  --border:2.5px solid var(--ink);
}
*{box-sizing:border-box}
body{margin:0;background:var(--paper)}
a{color:inherit;text-decoration:none}
a:hover{color:var(--indigo)}
.page{font-family:'Atkinson Hyperlegible',system-ui,sans-serif;font-size:18px;line-height:1.5;color:var(--ink);background:var(--paper);width:100%;overflow-x:hidden}
.display{font-family:'Bricolage Grotesque',sans-serif;font-weight:800;letter-spacing:-0.02em}
```

---

## 3. Global Interaction & Motion Rules

### 3.1 "Lift" cards and buttons (`.lift`)

```css
.lift{transition:transform .15s ease, box-shadow .15s ease}
.lift:hover{transform:translate(-3px,-3px);box-shadow:9px 9px 0 var(--ink) !important}
.btn:active{transform:translate(3px,3px);box-shadow:0 0 0 var(--ink) !important}
```
Apply `.lift` to: all buttons/CTAs, the three "How it works" cards, badge pills, the four accessibility tiles. Apply `.btn` to big pressable buttons (gives the squish-on-press feel). The `!important` is needed if shadows are set inline.

### 3.2 Focus (accessibility, required)

```css
button:focus-visible, a:focus-visible{outline:4px solid var(--ink);outline-offset:3px}
```

### 3.3 Floating stickers

```css
@keyframes bob{0%{transform:translateY(0) rotate(var(--r))}50%{transform:translateY(-9px) rotate(var(--r))}100%{transform:translateY(0) rotate(var(--r))}}
.sticker{animation:bob 3.6s ease-in-out infinite}
```
Each sticker sets its own tilt with `style="--r:7deg"` (or `-6deg`). Stagger with `animation-delay:1s` on the second.

### 3.4 Marquee ticker

```css
@keyframes slide{from{transform:translateX(0)}to{transform:translateX(-50%)}}
.track{animation:slide 34s linear infinite;display:flex;width:max-content}
```
The track contains **two identical groups** side by side (second copy gets `aria-hidden="true"`) so the loop is seamless at -50%.

### 3.5 Reduced motion (required)

```css
@media (prefers-reduced-motion: reduce){.track,.sticker{animation:none}}
```

---

## 4. Page Structure (top to bottom)

Root wrapper `.page` (fluid width, `overflow-x:hidden`). Sections in order:

1. Header
2. Hero (with scanner card)
3. Marquee ticker
4. How it works
5. Scam school (Learn)
6. Made for everyone (accessibility)
7. Family CTA
8. Footer

Anchor IDs: `#top` (hero), `#scan` (scanner card), `#learn`, `#access`, `#family`.

---

## 5. Section-by-Section Spec

### 5.1 Header
- Container: max-width 1240, padding `22px 20px`, `display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:16px 28px`.
- **Logo (left):** 46x46 rounded-square (radius 14) indigo tile, ink border, `3px 3px 0` shadow, white **shield-with-check** icon inside; next to it wordmark "[Brand Name]" (display font, 800, `1.55em`).
- **Nav (center):** links `Scan`, `Learn`, `Made for everyone`, `Family`. Bold, `1em`, anchor-link to sections, `padding:8px 2px`, wraps on small screens.
- **Text-size toggle (right):** small label "Text size" (`0.8em`, bold, muted) + a segmented control: white box, ink border, radius 16, padding 5, containing three buttons `A`, `A+`, `A++` (min 48x44, radius 11, display font 800). **Active** = ink background + lime text; **inactive** = white background + ink text. Button font sizes scale: `0.9em / 1.05em / 1.25em`. Each has `aria-label="Text size A"` etc.

### 5.2 Hero
- Section is `position:relative; overflow:hidden`. Two **flat decorative circles** behind content:
  - Lime circle `440px`, top-right (`right:-140px; top:-90px`).
  - Indigo-soft circle `300px`, bottom-left (`left:-110px; bottom:-80px`).
- Content container: `position:relative; max-width:1240; padding:36px 20px 84px; display:flex; flex-wrap:wrap; align-items:center; gap:56px 48px`. Two columns: **left copy** `flex:1 1 460px`, **right scanner** `flex:1 1 420px`; both `min-width:0`. Stacks on narrow screens.

**Left column:**
1. **Pill:** inline-flex, lime bg, ink border, radius 99, padding `8px 16px`, bold `0.85em`, with a small 4-point sparkle icon (filled ink) + text **"AI scam scanner + free scam school"**.
2. **H1:** "Spot the scam **before** it spots you." The word **"before"** sits in an inline-block lime highlight: lime bg, ink border, radius 12, padding `0 10px`, rotated `-1.5deg`.
3. **Paragraph:** "Paste a message, link or phone number. Our AI checks it in seconds and explains why in plain words, so you get safer every time you use it." (`1.25em`, `--body-2`, max-width ~34em).
4. **Buttons row** (flex, wrap, gap 16):
   - **Primary** "Scan something now" + right-arrow icon: indigo bg, white text, ink border, radius 18, min-height 60, padding `0 28px`, shadow `6px 6px 0`, display font 800 `1.2em`. Links to `#scan`.
   - **Secondary** "Take a 3-min lesson": white bg, ink text, same shape. Links to `#learn`.
5. **Trust row** (flex wrap, gap `12px 26px`, bold `0.95em`): three items, each = 28px ink circle with lime check icon + text: **"No jargon"**, **"Big-text mode"**, **"Made for the whole family"**.

**Right column: Scanner card (`#scan`)**, `position:relative; padding-top:18px`.
- **Two floating stickers** (class `.sticker`, z-index above card):
  - Coral sticker "**URGENT!! Act now**": top-right (`top:-6px; right:14px`), tilt `7deg`.
  - Yellow sticker "**You won a prize!**": bottom-left (`bottom:-22px; left:8px`), tilt `-6deg`, `animation-delay:1s`.
  - Both: padding `8px 16px`, ink border, radius 14, display font 800 `0.95em`, shadow `3px 3px 0`.
- **Card:** white, ink border, radius 28, shadow `10px 10px 0`, padding 26.
  - **Title row:** "Is this a scam?" (display, `1.5em`) + right pill "Live example" (indigo-tint bg, indigo-text color, `0.75em`, bold, radius 99).
  - **Tabs** (`role="tablist"`): three buttons "Message", "Link", "Phone", each with a stroke icon (chat bubble, chain link, phone handset) + label. Each `flex:1 1 110px`, min-height 52, ink border, radius 16, bold `0.95em`. **Active** = ink bg + lime text; **inactive** = white bg + ink text.
  - **Input display:** small bold muted label (changes per tab) then a **dashed** ink border box (`2.5px dashed`), radius 16, paper bg, padding `16px 18px`, min-height 96, showing the example text. This is display-only (a div), not a real input in the preview.
  - **"Scan with AI" bar:** full-width, min-height 58, ink bg, lime text, radius 16, display font 800 `1.1em`, sparkle icon on the left.
  - **Verdict panel:** margin-top 18, padding 18, ink border, radius 20, background = tint of verdict (coral-tint or yellow-tint).
    - Top row: verdict **chip** (bg coral or yellow, ink border, radius 99, display 800) + right label "Risk level".
    - **Risk meter:** height 20, white bg, ink border, radius 99, `overflow:hidden`; inner fill width = score %, same color as chip, with `border-right: 2.5px solid ink`.
    - **Flags list:** three rows, each = 28px ink circle with white number + sentence with a **bold lead phrase**.

**Example data for the three tabs (placeholders):**

| Tab | Label | Example text | Verdict | Meter | Color / tint |
|---|---|---|---|---|---|
| Message | Message you received | "URGENT: Your parcel is held at customs. Pay the small release fee today at bit-ly-parcel-id.co or it will be sent back." | Scam likely | 92% | coral / coral-tint |
| Link | Link you were sent | `https://secure-bank-login.verify-account.top/id` | Dangerous link | 96% | coral / coral-tint |
| Phone | Caller or sender | `+62 811 0000 4421: "Hi Mom, this is my new number. Please save it."` | Suspicious | 68% | yellow / yellow-tint |

Flags per tab (bold lead + rest):
- Message: **Fake urgency.** Real couriers do not threaten you with deadlines. / **Odd link.** The web address imitates a delivery brand. / **Surprise fee.** Fees are never collected through random links.
- Link: **Look-alike name.** It says bank, but the real domain is something else. / **Strange ending.** Banks rarely use cheap, unusual web endings. / **Login request.** Never sign in from a link you did not ask for.
- Phone: **Unknown number.** Claims to be family but is not in your contacts. / **Money may follow.** This opener is often the first step of a request. / **Safe move.** Call the old number and ask a question only they know.

### 5.3 Marquee Ticker
- Full-width band: ink bg, `padding:20px 0`, `overflow:hidden`, ink top/bottom border.
- Items (display font 700, `1.5em`, white, `white-space:nowrap`), each followed by a **lime 4-point sparkle** icon (20px), with `gap:22px`:
  **Fake prizes, Bank phishing, Parcel scams, Romance scams, Fake job offers, Investment traps, Impersonation, Fake invoices.**
- Duplicate the group once for the seamless loop (see 3.4).

### 5.4 How It Works
- Container padding `96px 20px 24px`, max-width 1240.
- Eyebrow "How it works" + H2 **"Three steps. Zero panic."**
- Three cards in a flex-wrap row (`gap:28px`, each `flex:1 1 280px`): white, ink border, radius 26, shadow `6px 6px 0`, padding 28, class `.lift`.
  - Top: 64x64 radius-20 number tile with ink border, display font 800 `1.9em`: **1** on indigo (white text), **2** on lime, **3** on yellow.
  - H3 (`1.5em`) + paragraph (`--body-2`):
    1. **Paste it in.** Got a strange text, link or call? Drop it in. No account needed to try.
    2. **AI takes a look.** It checks wording, links and pressure tactics, then gives a clear verdict.
    3. **Learn why.** Every red flag is explained, so next time you spot it yourself.

### 5.5 Scam School (`#learn`)
- Eyebrow "Scam school", H2 **"Learn the tricks, not just the verdict."**, paragraph "Short, friendly lessons built from real scam messages. Five minutes is plenty."
- A **bento grid** made with flex-wrap, `gap:28px`. Cards (all ink border, radius 28, padding 30, hard shadow `8px 8px 0`):

**a) Red flag of the week** (`flex:2 1 520px`, white card)
- Header: coral pill "Red flag of the week" + muted bold text `The "new number" trick`.
- **Chat bubble** (paper bg, ink border, radius `22px 22px 22px 6px`, padding `22px 24px`, `1.15em`, line-height 1.9) containing the text with three inline highlights (each: padding `2px 8px`, radius 8, `2px solid ink` border):
  "Hi Mom, my phone broke. This is my **[new number]** (yellow). Can you send money **[right now]** (coral)? **[Don't call]** (lime), I'm in a meeting."
- Below: three numbered explanation items (30px colored circles matching highlight colors):
  1. **Unknown number.** Real family can be reached on the old one too.
  2. **Pressure.** Scammers hate you taking time to think.
  3. **No calls.** They avoid the one thing that proves who they are.

**b) Quick lessons** (`flex:1 1 320px`, **ink card**, white text, shadow in **indigo** `8px 8px 0 #4A44EE`)
- Header: 44px lime rounded tile with book icon + H3 "Quick lessons".
- List of 4 lessons, each = title (bold) + right-aligned lime duration, then a 12px progress bar (track `--ink-soft`, fill lime, radius 99):
  Fake parcel fees (4 min, 100%) / The new number trick (5 min, 60%) / Spotting a fake link (3 min, 25%) / Too-good-to-be-true investments (6 min, 5%).

**c) Quiz: "Real or scam?"** (`flex:1 1 380px`, **lime card**, shadow `8px 8px 0 ink`)
- Eyebrow "Quick quiz" (uppercase), H3 "Real or scam?"
- White bubble: "Congrats! You won a free phone. Claim within 10 minutes: bit.ly/win-now"
- Two big buttons side by side: **"It is real"** (white bg, ink text) and **"It is a scam"** (ink bg, lime text). Both `.btn .lift`, min-height 58, radius 16, display 800 `1.1em`, shadow `5px 5px 0`.
- After answering, replace buttons with a white feedback box:
  - Correct (scam): **"Nailed it!"** / Wrong (real): **"Almost! It is a scam."**
  - Body for both: "A fake deadline and a shortened link are two classic red flags. Real prizes never rush you."
  - "Try another" button (ink bg, white text, min-height 44, radius 12) resets the quiz.

**d) Badges** (`flex:1 1 300px`, white card)
- H3 "Collect your badges", paragraph "Finish lessons, earn stickers, challenge friends and family."
- Pill badges (padding `10px 16px`, radius 99, ink border, shadow `4px 4px 0`, class `.lift`): **Link Sleuth** (indigo, white text), **Calm Caller** (yellow), **Parcel Pro** (lime), **Fake Prize Buster** (coral).

### 5.6 Made for Everyone (`#access`)
- Two-column flex-wrap. Left (`flex:1 1 340px`): eyebrow "Made for everyone", H2 **"Cool enough for your feed. Clear enough for your grandparents."**, paragraph "Big buttons, strong contrast and plain words by default. Turn on extras whenever you like."
- Right (`flex:1.4 1 460px`): four tiles in flex-wrap (`gap:20px`, each `flex:1 1 200px`), radius 24, ink border, shadow `6px 6px 0`, padding 24, `.lift`, each with a 34px stroke icon, H3 (`1.3em`) and `0.95em` text:
  1. **Big-text mode**: indigo-soft bg, "A" glyph icon. "Three sizes, one tap, from the top of every page."
  2. **Read aloud**: yellow bg, speaker icon. "Hear every verdict and lesson spoken in calm, simple words."
  3. **Night mode**: ink bg, white text, lime moon icon, indigo shadow. "Easy on the eyes for late-night doom-scrolling."
  4. **Plain words**: lime bg, chat icon. "No tech talk. If a sentence needs a dictionary, we rewrite it."

### 5.7 Family CTA (`#family`)
- Container padding `88px 20px 96px`. One big panel: **indigo bg, white text**, ink border, radius 36, shadow `10px 10px 0`, padding `52px 40px`, `position:relative; overflow:hidden`.
- Decor: lime circle 240px top-right (`right:-70px; top:-70px`) and yellow circle 120px (`right:120px; bottom:-60px`).
- Content (max-width 700, relative): H2 **"Got a weird message right now?"**; paragraph "Scan it first. Then send the family guide to someone you love, so the whole group stays one step ahead." (`--on-indigo-text`, `1.2em`).
- Buttons: **"Scan it free"** (lime bg, ink text) and **"Share the family guide"** (white bg, ink text). Same shape/shadow as hero buttons.

### 5.8 Footer
- White bg, top border `2.5px solid ink`. Container padding `32px 20px`, flex wrap, space-between.
- Wordmark "[Brand Name]" (display 800 `1.3em`), links (Scan, Learn, Accessibility, Family guide; bold `0.95em`), and small muted note.

---

## 6. Icons

Use **inline SVG, stroke style** (no emoji, no icon fonts). Default attributes: `viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"` to `2.6`, `stroke-linecap="round" stroke-linejoin="round"`. Decorative icons get `aria-hidden="true"`.

Icons needed: shield-with-check (logo), chat bubble, chain link, phone, right arrow, check, book, speaker, moon, "A" type glyph, filled 4-point sparkle. Paths used in the preview (24x24 grid):

```
shield+check: M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6l7-3z   +   M9 12l2 2 4-4
chat:         M4 5h16v11H9l-5 4V5z
link:         M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1   +   M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1
phone:        M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A15 15 0 0 1 3 6a2 2 0 0 1 2-2z
arrow-right:  M5 12h14M13 6l6 6-6 6
check:        M5 12l5 5 9-10
book:         M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2V5z   +   M8 7h7
speaker:      M4 10v4h4l5 4V6L8 10H4z   +   M16 9a4 4 0 0 1 0 6
moon:         M20 14.5A8 8 0 0 1 9.5 4 8 8 0 1 0 20 14.5z
type "A":     M5 19L11 5l6 14M7.5 14h7
sparkle(fill):M12 2l2.6 7.4L22 12l-7.4 2.6L12 22l-2.6-7.4L2 12l7.4-2.6z
```

---

## 7. Responsive Behavior

This is a **fluid page**, not fixed width.
- Use `flex-wrap: wrap` with `flex: 1 1 <basis>` and `min-width: 0` rather than media-query-heavy grids. Columns stack automatically when they can't fit.
- Max content width 1240 centered; side gutter `20px` at all sizes (no horizontal scrolling).
- Headlines use `clamp()` so they shrink gracefully on phones.
- Nav and toolbar wrap onto extra lines on small screens.
- Floating stickers stay inside the card's bounds (use small offsets like `left:8px`, `right:14px`) so they never cause horizontal overflow.
- The decorative circles are clipped by `overflow:hidden` on their section.
- Verify at **360px, 768px, 1280px+**.

---

## 8. Accessibility Requirements (do not skip)

- Real elements: `<button>` for actions, `<a href>` for links, `<nav aria-label="Main">`, `<header>`, `<section>`, `<footer>`. No clickable `div`s.
- Tabs: container `role="tablist"`, each button `role="tab"` with `aria-selected` (true on the active one) and `aria-controls` pointing to the result panel.
- Text-size buttons have `aria-label`; mark the active one with `aria-pressed="true"`.
- Minimum tap target **44x44px** (most are 48 to 60).
- Keep contrast per section 2.1. Never convey meaning by color alone (the verdict chip always has text; the meter has a text label).
- Visible `:focus-visible` outlines (section 3.2).
- Honor `prefers-reduced-motion`.
- Decorative SVG/circles: `aria-hidden="true"`. The duplicated marquee group: `aria-hidden="true"`. Give the marquee section `aria-label="Scams we help you spot"`.
- Set `<html lang="en">` (change if copy is translated).
- Quiz feedback region: `aria-live="polite"` so the result is announced.

---

## 9. Interactions to Implement (vanilla JS)

Everything else is static. Keep JS small and dependency-free.

### 9.1 Text-size toggle
Three sizes set on the `.page` wrapper's `font-size`:
```js
const SIZES = { A: 18, 'A+': 20.5, 'A++': 23 };
document.querySelectorAll('[data-size]').forEach(btn => {
  btn.addEventListener('click', () => {
    page.style.fontSize = SIZES[btn.dataset.size] + 'px';
    document.querySelectorAll('[data-size]').forEach(b => {
      const on = b === btn;
      b.setAttribute('aria-pressed', on);
      b.classList.toggle('is-active', on);   // active = ink bg + lime text
    });
    try { localStorage.setItem('textSize', btn.dataset.size); } catch(e){}
  });
});
```
Restore saved size on load (inside try/catch). Default is `A`.

### 9.2 Scanner tabs (mock)
Three data objects (see 5.2 table). On tab click: set active tab styles, then update the label, example text, verdict chip text + color, verdict panel tint, meter width, and the three flags. Meter width can transition (`transition: width .4s ease`).
```js
const TABS = [ /* {label, input, verdict, meter, color, tint, flags:[{title,text}]} x3 */ ];
function render(i){ /* write fields into the DOM; set chip/meter bg to color; panel bg to tint */ }
```

### 9.3 Quiz
State: `null | 'real' | 'scam'`. Show the two buttons when `null`. On pick, hide buttons and show the matching feedback box (correct if `'scam'`, otherwise "Almost!"). "Try another" returns to `null`.

### 9.4 Pure CSS effects
Hover lift, press squish, floating stickers, and the marquee need no JS.

---

## 10. Suggested File Structure (plain project)

```
/index.html
/css/styles.css     (tokens, base, components, sections, responsive)
/js/main.js         (text-size toggle, scanner tabs, quiz)
/assets/            (optional: favicon only; the design uses no photos)
```
If the project already has a structure or stack, **follow it** and keep these class names and tokens as the design language. Use semantic class names such as `.btn`, `.btn--primary`, `.btn--secondary`, `.card`, `.pill`, `.sticker`, `.tab`, `.tab.is-active`, `.eyebrow`, `.bento`.

---

## 11. Implementation Order (recommended)

1. Add fonts, tokens, base styles, and the global helper classes (`.lift`, `.btn`, `.sticker`, focus, reduced motion).
2. Build header + text-size toggle.
3. Build hero (left copy first, then scanner card with static "Message" example).
4. Add marquee, then How it works.
5. Build the Scam school bento cards.
6. Build Made for everyone, Family CTA, footer.
7. Add JS (text size, tabs, quiz).
8. Responsive pass at 360 / 768 / 1280px, then run the checklist below.

---

## 12. Acceptance Checklist

- [ ] Fonts load: Bricolage Grotesque for headings/buttons, Atkinson Hyperlegible for body.
- [ ] Every card/button/pill has a `2.5px` ink border and a **hard** (zero-blur) offset shadow.
- [ ] Colors match the tokens exactly; no gradients, no blurry shadows, no emoji icons.
- [ ] Hover lifts cards/buttons up-left and grows the shadow; pressing squishes them to flat.
- [ ] Hero H1 has the tilted lime highlight on "before"; two bobbing stickers appear around the scanner card.
- [ ] Marquee loops seamlessly and stops when reduced motion is on.
- [ ] Text-size toggle visibly scales the **entire** page and persists across reloads.
- [ ] Scanner tabs swap text, verdict chip, meter width/color, panel tint, and flags.
- [ ] Quiz shows correct/incorrect feedback and can be reset.
- [ ] Layout has no horizontal scroll at 360px; columns stack cleanly.
- [ ] All tap targets are at least 44px; keyboard focus rings are visible; tabs have ARIA roles.
- [ ] Copy matches this document (placeholder brand name "[Brand Name]" can be swapped globally).

---

## 13. Notes and Honest Limits

- The scanner is a **visual mock**. Verdicts, percentages, lesson progress, and badges are sample content, not real analysis or real user data. Wire them to a real backend later if desired.
- "Read aloud" and "Night mode" appear as feature tiles only; they are not implemented in the preview. Add them later if wanted (e.g. Web Speech API for read-aloud, a dark token set for night mode).
- Keep the brand name, copy, and example scam texts easy to replace; all are placeholders.
