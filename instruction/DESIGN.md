# Design System Reference

> A visual and interaction reference for a distinctive, approachable, and engaging website.
>
> **Core principle:** One identity. Multiple experiences. No visitor left behind.

This document defines the visual language, interaction principles, accessibility expectations, and design constraints. It intentionally does not define page layouts, page copy, content structure, or implementation-specific component placement.

---

## 1. Design Direction

The design should feel like a carefully crafted digital environment: technical, atmospheric, expressive, and slightly playful without becoming difficult to use.

It should balance three qualities:

- **Distinctive:** recognizable through its visual details rather than generic templates.
- **Professional:** clear, deliberate, consistent, and suitable for a broad audience.
- **Inviting:** rewarding to explore, with small moments of personality and interaction.

The interface may be expressive, but the information must remain easy to understand. Decoration should support the experience, never compete with it.

## 2. Core Design Principles

### Character with clarity
Keep the interface's recognizable technical visual language, but make every important element understandable to a first-time visitor.

### Exploration without confusion
Interactions should invite curiosity. Visitors should not need to guess how to navigate, discover content, or return to a familiar state.

### Motion with purpose
Animation should communicate state, provide feedback, or add atmosphere. Avoid motion that delays tasks or distracts from reading.

### Consistency over novelty
New visual effects and interactive details must follow the same visual rules. Do not introduce a new style for every feature.

### Accessible by default
Readability, keyboard access, clear focus states, and reduced-motion support are foundational requirements, not optional enhancements.

## 3. Visual Language

### Overall aesthetic
- Dark, atmospheric surfaces form the primary visual foundation.
- Use restrained technical details such as fine lines, subtle grids, panel boundaries, small status indicators, and precise alignment.
- Warm accent colors provide focus and personality.
- Secondary cool tones may be used sparingly to distinguish supporting information or states.
- Pixel-inspired or character-based details may add charm, provided they remain secondary to the content.
- Avoid excessive glow, blur, gradients, noise, or decorative clutter.

### Color roles
Use semantic color roles rather than choosing colors independently for each component.

| Role | Purpose |
|---|---|
| Primary background | Main dark canvas |
| Elevated surface | Panels, cards, and interactive regions |
| Primary text | Main readable content |
| Secondary text | Supporting information and metadata |
| Primary accent | Important actions, active states, and key highlights |
| Secondary accent | Supporting emphasis and visual variation |
| Border | Separation, framing, and technical detailing |
| Success / warning / error | Clear status communication |

Color requirements:
- Maintain strong text-to-background contrast.
- Never communicate meaning through color alone.
- Keep accent colors intentional and limited.
- Ensure hover, focus, active, disabled, and selected states remain distinguishable.
- Avoid using low-contrast decorative text for information visitors need.

Exact color values should be sampled from and validated against the approved visual reference before being formalized as implementation tokens.

### Typography
- Use a highly readable primary typeface for body text and important information.
- A technical or monospace typeface may be used for labels, metadata, identifiers, and small interface details.
- Keep decorative or display typography limited to short headings.
- Establish a clear hierarchy through size, weight, spacing, and contrast.
- Avoid making essential information depend on tiny uppercase labels or condensed lettering.
- Body text should remain comfortable to read on mobile and desktop.

### Borders, shapes, and surfaces
- Prefer fine, deliberate borders over heavy containers.
- Use consistent corner treatments; do not mix unrelated shape languages.
- Panels should feel related through shared spacing, border, and surface rules.
- Use shadows and glow sparingly, primarily to communicate depth or interaction.
- Avoid turning every element into a card or pill.

### Icons and decorative graphics
- Use a consistent icon style and stroke weight.
- Icons that perform actions must have an accessible name or visible label.
- Decorative graphics must not resemble controls.
- Pixel art and character illustrations should add personality without reducing legibility or obscuring important information.

## 4. Two Presentation Experiences

The website should offer two presentation modes based on visitor preference, **not age or generation**. People of any age may prefer either experience.

Both modes belong to one design system and must communicate the same underlying information.

### Explore Mode
An expressive, immersive presentation for visitors who enjoy discovering details and interacting with the interface.

- Retain the full visual character and technical atmosphere.
- Allow richer transitions, character moments, visual feedback, and optional discoveries.
- Use playful micro-interactions and meaningful interactive details.
- Keep navigation understandable and controls discoverable.
- Ensure visual effects never obstruct essential content or actions.

### Essential Mode
A calmer, more direct presentation for visitors who prefer clarity and minimal interaction.

- Reduce decorative density and nonessential motion.
- Prioritize readable typography, clear labels, and straightforward controls.
- Simplify visual emphasis without removing meaningful information.
- Keep the same content, destinations, and core functionality.
- Preserve the overall visual identity through color, typography, borders, and restrained technical details.

### Mode-switching rules
- Offer a simple, neutral choice between the two experiences.
- Do not label modes by age, generation, ability, or technical skill.
- Make the choice easy to understand without a long explanation.
- Allow visitors to switch modes at any time.
- Preserve the visitor's choice between visits only when appropriate, and provide an obvious way to change it.
- Switching modes must not reset progress, discard input, or send visitors to an unrelated place.
- Respect system reduced-motion preferences in both modes.
- Essential Mode must never feel like a lesser, incomplete, or second-class version.

## 5. Interaction and Engagement

The website should feel engaging because it responds thoughtfully, not because it constantly demands attention.

### Good interaction patterns
- Clear hover, focus, pressed, selected, and disabled states.
- Small transitions that explain what changed.
- Optional details that visitors can reveal when curious.
- Character or visual reactions that reinforce an action without blocking it.
- Interactive elements with predictable behavior and clear outcomes.
- Easy ways to close, dismiss, undo, or return when an interaction changes the interface.

### Interaction constraints
- Do not hide essential information behind unexplained gestures.
- Do not make hover the only way to discover important controls.
- Avoid surprise audio, forced animations, flashing effects, and unnecessary pop-ups.
- Do not make visitors wait through decorative transitions before they can continue.
- Avoid fake controls, misleading status indicators, or interactions that do nothing.
- Provide feedback for actions, including loading, success, and error states where relevant.

### Motion principles
- Prefer short, purposeful transitions.
- Use motion to communicate continuity, hierarchy, and state changes.
- Avoid excessive parallax, looping movement, and constant ambient animation.
- Keep important content stable while visitors read or interact.
- Support `prefers-reduced-motion`; provide a genuinely calmer experience rather than merely slowing every animation.
- Never use flashing or rapidly alternating effects.

## 6. Usability and Accessibility Improvements

These requirements apply to every presentation mode.

### Readability
- Keep body text comfortably sized across viewport sizes.
- Maintain adequate line height and paragraph spacing.
- Ensure metadata remains legible and does not carry essential meaning alone.
- Avoid rotated text for important navigation labels.
- Do not rely on tiny uppercase text to communicate important information.

### Navigation
- Use recognizable labels and consistent navigation behavior.
- Make the current location or active section clear.
- Provide visible focus indicators for keyboard users.
- Ensure interactive targets are large enough and comfortably spaced, especially on touch screens.
- Tooltips may supplement labels, but must not replace necessary instructions.

### Keyboard and assistive technology
- All functionality must be usable with a keyboard.
- Use semantic controls and meaningful accessible names.
- Preserve a logical focus order.
- Ensure focus is visible and not obscured.
- Provide text alternatives for meaningful images and hide purely decorative images from assistive technology.
- Announce important dynamic status changes accessibly.

### Responsive behavior
- Support narrow phones, larger phones, tablets, laptops, and wide displays.
- Reflow content rather than forcing desktop interactions onto mobile.
- Avoid horizontal scrolling except where a genuinely horizontal interaction is intentional and has an accessible alternative.
- Ensure text, controls, and interactive targets remain usable at increased zoom.
- Test both presentation modes at multiple viewport sizes.

### Performance
- Keep animation and visual effects lightweight.
- Avoid loading large decorative assets before essential content.
- Prefer progressive enhancement: the website should remain usable if an optional effect or animation fails.
- Avoid interactions that depend on fragile timing or excessive scripting.

## 7. Design Tokens and Consistency

Formalize repeated visual decisions as shared tokens rather than one-off values.

Recommended token groups:
- Color roles
- Typography scale and font roles
- Spacing scale
- Border widths and colors
- Corner radii
- Elevation and shadow levels
- Motion durations and easing
- Focus-ring treatment
- Interactive state treatments

Use a small, coherent scale. New components should reuse established tokens unless there is a clear reason to extend the system.

## 8. Preserve and Avoid

### Preserve
- The distinctive technical atmosphere.
- The balance of dark surfaces, precise detailing, and restrained warm accents.
- The expressive character of the interface.
- Small moments of charm and discovery.
- A professional level of visual consistency.
- The freedom to explore without sacrificing clarity.

### Avoid
- Generic template styling that removes the site's character.
- Excessive decoration or effects without a functional purpose.
- Tiny, low-contrast text.
- Navigation that depends on guessing or prior technical knowledge.
- Age-based assumptions about visitor preferences.
- Different information or capabilities between presentation modes.
- Making the simpler mode look unfinished.
- Replacing the established visual language with an unrelated design trend.

## 9. Quality Checklist

Before approving a visual or interaction change, verify:

- [ ] Does it preserve the established visual identity?
- [ ] Is its purpose clear to a first-time visitor?
- [ ] Is important information easy to read?
- [ ] Are interactive and decorative elements distinguishable?
- [ ] Does it work with keyboard and touch input?
- [ ] Are focus, hover, active, and disabled states clear?
- [ ] Does it respect reduced-motion preferences?
- [ ] Does it work in both presentation modes?
- [ ] Does it preserve equivalent information and core functionality?
- [ ] Does it remain usable on narrow and wide screens?
- [ ] Does the change add value without unnecessary visual noise?

---

**Design philosophy:** One identity. Multiple experiences. No visitor left behind.
