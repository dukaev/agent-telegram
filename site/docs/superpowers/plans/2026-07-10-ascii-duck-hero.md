# ASCII Duck Hero Implementation Plan

**Goal:** Replace the hero Live User Session mockup with a responsive Canvas 2D ASCII rendering of `vid.mp4`.

**Architecture:** An app-specific landing renderer owns a hidden optimized video element, a low-resolution offscreen sample canvas, and a visible glyph canvas. The renderer draws outside React state, pauses offscreen, and obeys reduced motion. The existing Scenario Lab continues to demonstrate the product UI lower on the page.

**Verification tier:** Tier 3.

## Task 1: Source preparation and pure mapping utilities

**Files:**
- Create: `public/duck-laptop.mp4`
- Create: `src/landing/ascii-frame.ts`
- Create: `src/landing/ascii-frame.test.ts`

- Generate a silent, web-optimized 640 px derivative from the supplied video.
- Implement pure glyph-strength, background-rejection, and color-lift helpers.
- Unit-test white rejection, dark-outline visibility, saturated-yellow density, and responsive grid bounds.

## Task 2: ASCII video renderer

**Files:**
- Create: `src/landing/components/ascii-duck-video.tsx`

- Decode the static source through a muted inline video element.
- Sample into a 54–92 column offscreen canvas.
- Draw colored glyphs into the visible canvas at no more than 12 fps.
- Pause decode/render work offscreen and cancel scheduled frames on unmount.
- Draw a stable first frame for reduced-motion users.

## Task 3: Hero composition and visual system

**Files:**
- Modify: `src/landing/components/hero.tsx`
- Modify: `src/styles.css`

- Remove the hero `TelegramDemo`, bot scenario dependency, live-session badges, and cursor tilt.
- Compose the ASCII renderer as the primary hero visual.
- Add restrained terminal framing, scanlines, vignette, glow, and responsive sizing.
- Keep the entrance reveal and existing conversion copy/actions intact.

## Task 4: Acceptance and decision evidence

**Files:**
- Modify: `e2e/landing.spec.ts`
- Modify: `docs/toolcraft/agent-worklog.md`

- Assert the ASCII canvas replaces the old hero demo.
- Assert autoplay/frame progress, reduced-motion pause, and mobile containment.
- Record the video reference study, renderer choice, pipeline, verification tier, skipped full perf gate, and risks.

## Task 5: Verification

- Run `npm run verify:quick`.
- Run `npm run build`.
- Run landing-focused Playwright tests.
- Inspect desktop and mobile output in an agent-controlled browser.
- Confirm no console errors, no horizontal overflow, and visible frame progression while the hero is onscreen.

## Task 6: Higher-detail rendering pass

**Files:**
- Modify: `src/landing/ascii-frame.ts`
- Modify: `src/landing/ascii-frame.test.ts`
- Modify: `src/landing/components/ascii-duck-video.tsx`
- Modify: `docs/superpowers/specs/2026-07-10-ascii-duck-hero-design.md`
- Modify: `docs/toolcraft/agent-worklog.md`

- Increase the responsive grid to a 72–144-column range without changing source video resolution.
- Expand the glyph ramp so midtone and halftone differences survive the mapping pass.
- Grade monochrome strokes through a warm strength-dependent palette.
- Keep the 12 fps cadence, offscreen suspension, and reduced-motion still behavior unchanged.
- Update unit expectations and run targeted TypeScript, Vitest, Playwright, production build, and agent-browser visual/cadence checks.
