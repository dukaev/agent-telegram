# ASCII Duck Hero Design

## Objective

Replace the hero's `Live User Session` terminal/chat mockup with an autonomous ASCII rendering of the supplied laptop-duck video. The result should feel like a living brand character rather than another product screenshot while preserving the landing's dark technical tone.

## Verification Note

Verification tier: Tier 3 — custom media renderer and autonomous animation.

Reason: This pass adds a video source, per-frame pixel sampling, Canvas 2D glyph rendering, responsive sizing, reduced-motion behavior, and a new hero composition. It does not change the copied Toolcraft runtime, schema, editor surfaces, persistence, export, or the product's deeper interactive scenario lab.

Run: `npm run verify:quick`, `npm run build`, landing-focused Playwright acceptance, and an agent-controlled browser review at desktop/mobile widths with console and frame-progress checks.

Skip: The full Toolcraft performance suite is not required for this post-first-working landing feature. A targeted animation workload check covers the changed renderer.

## Video Reference Study

- `referenceLocation`: `/Users/aslan/Documents/agent-telegram-site/vid.mp4`
- `extractionEvidence`: The refreshed source was inspected again after replacement. `ffprobe` reports H.264 video at 1388×1492, 24 fps, 4.041667 seconds, with AAC audio. A one-second contact sheet was extracted to `/tmp/agent-telegram-video-refresh.r04VaH/contact-sheet.jpg`.
- `storyboard`:
  - 00:00 — monochrome line-art duck sits with an open black laptop; small sparkle marks flank the character.
  - 00:01 — body and laptop remain anchored while the duck's right wing shifts across the keyboard.
  - 00:02 — the wing releases downward and the body subtly settles.
  - 00:03 — the wing returns toward the keyboard, completing the restrained work loop.
- `transitionAnalysis`: The camera, baseline, body, laptop, and framing remain stable. Motion is concentrated in the right wing with subtle body settling; sparkle marks stay fixed. The refreshed source is portrait-oriented, monochrome, and has a near-white background plus embedded audio. The background and audio are intentionally omitted from product output, while dark line work is lifted to warm amber for contrast.
- `behaviorDecomposition`: decode the video; sample a small responsive grid; reject near-white low-saturation pixels; derive glyph density from saturation and inverse luminance; draw colored monospace glyphs to a transparent Canvas 2D surface; cap rendering at 12 fps; pause video and frame work when offscreen; show a deterministic still frame for reduced motion.
- `acceptanceMapping`:
  - `landing-ascii-hero`: hero exposes an ASCII canvas and no longer exposes the Acme Test Bot / Live User Session panel.
  - `landing-ascii-progress`: visible autoplay advances the source and redraws the canvas without console errors.
  - `landing-ascii-reduced-motion`: reduced motion holds a static ASCII frame instead of autoplaying.
  - `landing-ascii-mobile`: the canvas stays inside the viewport at 390 px without horizontal overflow.

## Product Decisions

- Visible output: one warm-amber ASCII line-art duck on a transparent near-black hero surface, with restrained scanline/glow treatment and a small technical source label.
- Media flow: ship an optimized, silent derivative of `vid.mp4` in `public/`; the original remains the reference input at the workspace root.
- Controls: none. This is autonomous decorative motion with no user transport, duration editing, or export.
- Timeline: none. The source video owns its six-second loop; reduced motion pauses it.
- Layers: none. Video decoding, offscreen sampling, Canvas glyph output, and CSS frame treatment are implementation passes, not user-editable product entities.
- Persistence/settings transfer: none.
- Export: none; the landing consumes the renderer directly.

## Renderer Technique

- `sourceRepresentation`: optimized 640×688 H.264 video derived from the refreshed supplied 1388×1492 source, without audio.
- `productRepresentation`: colored ASCII glyph field with near-white background removal.
- `previewRenderer`: Canvas 2D output plus a tiny offscreen Canvas 2D sampling buffer.
- `exportRenderer`: none.
- `rendererWorkload`: responsive 54–92-column sample grid, approximately 1,600–4,700 sampled cells per drawn frame, capped at 12 fps.
- `rendererStrategy`: Canvas 2D.
- `whyNotAlternativeStrategies`: DOM spans would create thousands of nodes and repeated layout; WebGL would add shader/program complexity for a bounded grid below 4,700 cells; preprocessing every frame into a baked asset would remove responsive glyph sizing and color treatment.
- `fidelityRisks`: white-background rejection can remove very pale highlights; a source shadow near the loop end can temporarily reduce glyph brightness.
- `performanceRisks`: `getImageData` is synchronous, so work is capped, sampled at low resolution, paused offscreen, and kept outside React state updates.

## Render Pipeline Inventory

1. Video decode — browser media pipeline; cache key is the static optimized source; invalidated only by source load/seek.
2. Sample pass — offscreen Canvas 2D at the current glyph grid; invalidated by video time or responsive column count.
3. ASCII mapping — main thread typed-array scan; converts pixels into glyph, color, and opacity values; invalidated with the sample pass.
4. Product draw — visible Canvas 2D; invalidated by mapped frame or output resize.
5. CSS composite — border, vignette, scanlines, and glow; compositor-owned and independent from video pixels.

## Accessibility

- The canvas has a descriptive image label.
- The source video is muted, hidden from assistive technology, and not focusable.
- `prefers-reduced-motion` pauses the video and renders one stable frame.
- If media autoplay fails, the loaded first frame remains visible.

## Detail Pass

Verification tier: Tier 3 — the renderer's per-frame workload and output fidelity change.

- Goal: make the refreshed monochrome line-art source read as a finer illustration rather than a coarse ASCII silhouette.
- Renderer change: increase the responsive grid from 54–92 columns to 72–144 columns and expand the density ramp from 10 to 22 glyphs.
- Expected hero workload: approximately 125×70 sampled cells at the current desktop width; maximum 144×80, or 11,520 sampled cells per drawn frame.
- Color change: map neutral source strokes into a strength-dependent warm palette so thin halftone marks remain visible instead of collapsing into one outline color.
- Unchanged: 12 fps cap, offscreen pause, reduced-motion still, source media, React state boundary, controls, timeline, layers, persistence, and export.
- Performance decision: retain Canvas 2D because the hard limit remains 11,520 sampled cells with only non-background cells producing `fillText`; targeted browser cadence and console checks must pass at the real desktop output before delivery.
