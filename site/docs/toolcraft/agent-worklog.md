# Implementation Worklog

## Status

Mode: product

The public route is now a focused landing page for `agent-telegram`: a local-first Telegram interface for AI agents, led by bot testing and supported by workflow, inbox, and community-management use cases.

## Decision Trail

### Iteration 2 — Interactive agent-telegram landing page

- Request: Replace the starter experience with a React + Vite landing page using Motion and GSAP, with substantial interaction.
- User-visible result: A responsive product narrative with an animated bot-test demo, switchable scenarios, interactive use-case cards, a scroll-driven architecture signal, safety proof, copy-to-install actions, and mobile navigation.
- Source checked: The public `dukaev/agent-telegram` repository, including its README, commands, local IPC model, JSON receipts, authentication flow, and safety gates.
- Product position: “Telegram for AI agents.” The page does not claim that the current repository is a native MCP server; it presents the CLI as agent-ready infrastructure that can sit behind an MCP adapter or agent skill.
- Primary audience: Agent builders who need to exercise a Telegram bot as a real user instead of mocking Telegram updates.
- Primary conversion: Copy `npm install -g agent-telegram`; secondary conversion: open GitHub.
- Visual direction: Dark technical interface, restrained cyan/green signal color, oversized editorial typography, real command/output surfaces, and no decorative stock imagery.
- Motion direction: Motion handles component-state transitions and reveal choreography; GSAP handles the hero entrance and scroll-linked architecture signal. Reduced-motion users receive complete, static content without autoplay.
- Alternatives rejected: A generic “MCP server” pitch would overstate the implementation; a conventional feature-card SaaS page would fail to demonstrate the product; video-only storytelling would be less inspectable and less interactive.
- Compatibility decision: The existing Toolcraft runtime source remains in the repository, but the public route intentionally acts as a route-level marketing-site escape hatch rather than an editor surface.

### Iteration 3 — ASCII laptop-duck hero

- Request: Replace the hero Acme Test Bot / Live User Session mockup with the supplied `vid.mp4` rendered as ASCII, using the Toolcraft workflow.
- Task type: Tier 3 custom media renderer and autonomous animation on the public landing route.
- User-visible result: The hero now features a colored ASCII duck typing on a laptop inside a restrained technical frame; the detailed interactive Telegram demo remains lower in the Scenario Lab.
- Source/reference checked: The supplied video and its later same-name replacement were each inspected with `ffprobe` and one-second storyboard contact sheets before integration. The active source is a 4.041667-second, 1388×1492 monochrome line-art loop.
- Reference inputs: `/Users/aslan/Documents/agent-telegram-site/vid.mp4`; `/tmp/agent-telegram-video-refresh.r04VaH/contact-sheet.jpg`; full refreshed study in `docs/superpowers/specs/2026-07-10-ascii-duck-hero-design.md`.
- Docs/contracts read: `AGENTS.md`, Toolcraft workflow, runtime boundary, media upload, reference study, renderer technique, performance, brainstorming, writing-plans, and browser workflow.
- Contract rules applied: `video-reference-analysis`, `renderer-technique-inventory`, `performance-coverage-levels`, `timeline-mode-choice`, `layers-enable-only-when-needed`, and `workflow-required`.
- Decision: Use a silent 640×688 H.264 derivative, an offscreen 54–92-column sample canvas, and a visible Canvas 2D warm-glyph renderer capped at 12 fps. Treat the motion as autonomous decoration with no Toolcraft timeline, controls, persistence, layers, or export.
- Alternatives rejected: A raw video would retain the white rectangle and feel like an inserted asset; DOM glyph spans would create thousands of layout nodes; WebGL is unnecessary for fewer than 4,500 sampled cells; a baked ASCII video would lose responsive glyph sizing and the transparent product surface.
- State/output mapping: Browser video time feeds the offscreen sample canvas; sampled RGB values feed background rejection, glyph density, lifted color, and alpha; mapped cells draw into the visible hero canvas. Intersection state pauses decode/render work, and reduced-motion state seeks to and holds the first frame.
- Files changed: optimized public video asset, ASCII mapping utilities/tests, ASCII renderer component, hero composition, landing styles, browser acceptance, feature spec/plan, and this worklog.
- Verification: `npm run verify:quick`, `npm run build`, four ASCII mapping tests, full five-test landing Playwright acceptance, and the refreshed-source desktop/mobile/reduced-motion regression subset passed. Agent-controlled browser review confirmed the active 4.04-second source, correct portrait proportions, zero horizontal overflow, visible frame progression, and no console warnings or errors.
- Skipped checks: Full Toolcraft performance suite is not required for a post-first-working landing feature; targeted frame-progress, output-pixel, console, viewport, and mobile containment checks cover the changed workload.
- Risks: Synchronous `getImageData` can become expensive if the grid grows; bounds, 12 fps cadence, offscreen pausing, and zero React frame state keep the workload controlled. Near-white rejection intentionally removes pale source highlights.

### Iteration 4 — Higher-detail ASCII pass

- Request: Make the refreshed ASCII duck more detailed.
- Task type: Tier 3 renderer workload and visual-fidelity refinement.
- User-visible result: The hero duck now uses finer line work, a 22-step character ramp, and warm tonal separation; the controlled desktop viewport renders 129×72 glyph cells instead of roughly 76×43.
- Source/reference checked: The active refreshed `vid.mp4`, its one-second contact sheet, the previous coarse browser render, and the new controlled-browser render were compared.
- Reference inputs: `/Users/aslan/Documents/agent-telegram-site/vid.mp4`; `/tmp/agent-telegram-video-refresh.r04VaH/contact-sheet.jpg`; live output at `http://127.0.0.1:3002/`.
- Docs/contracts read: Toolcraft brainstorming, writing-plans, renderer technique, performance, browser workflow, and the existing ASCII feature spec/plan.
- Contract rules applied: `renderer-technique-inventory`, `performance-coverage-levels`, `video-reference-analysis`, and `workflow-required`.
- Decision: Increase the responsive grid to 72–144 columns, expand the density ramp to 22 ordered glyphs, and grade neutral strokes by strength through a warm three-channel palette. Preserve the 12 fps cap, source resolution, offscreen pause, reduced-motion still, and direct Canvas draw path.
- Alternatives rejected: Increasing source resolution cannot add line detail after the glyph grid downsample; DOM glyphs would add layout cost; 24 fps would increase work without improving the intentionally restrained motion; WebGL remains disproportionate for a maximum 11,520 sampled cells and sparse foreground `fillText` calls.
- State/output mapping: Responsive canvas width selects 72–144 columns; source aspect derives rows; pixel strength selects one of 22 glyphs and a warm tone; the visible canvas publishes column/row/frame observables used by browser acceptance.
- Files changed: ASCII mapping utility/test, ASCII video renderer, browser acceptance, feature spec/plan, and this worklog.
- Verification: `npm run verify:quick` passed with 21 tests total (12 Node and 9 Vitest); `npm run build` passed; three targeted Playwright scenarios passed. Agent-controlled browser measured 129×72 cells at 1265 px viewport width, 0 px overflow, 0 console warnings/errors, and 0.65 seconds of source advance over a 0.6-second observation.
- Skipped checks: Full Toolcraft performance suite is not required for a post-first-working localized renderer refinement; the real source, maximum declared bounds, targeted acceptance, cadence, overflow, console, and visual checks cover the changed path.
- Risks: Maximum work rose from 4,692 to 11,520 sampled cells per drawn frame. The effect stays capped at 12 fps and background cells do not issue `fillText`; future density increases should trigger WebGL or baked-frame evaluation rather than another Canvas 2D increase.

### Iteration 5 — Prompt-first QR onboarding

- Request: Make the first command an agent prompt instead of npm install, reduce setup to scanning one QR code, then commit and push everything.
- Task type: Tier 2 conversion-flow behavior plus commit-ready delivery.
- User-visible result: The hero's first interactive action copies a complete setup prompt; step two says to scan one Telegram QR code. The final CTA repeats the same two-step flow, while npm install remains only as a small manual fallback.
- Source/reference checked: Existing `agent-telegram` repository findings for npm installation, local QR authentication, real-user sessions, read-only checks, and explicit confirmation gates; current desktop and final-CTA browser renders.
- Reference inputs: Existing landing at `http://127.0.0.1:3002/`; no new external design assets.
- Docs/contracts read: Toolcraft brainstorming, writing-plans, workflow, existing landing spec/plan, GitHub publish workflow, and browser workflow.
- Contract rules applied: `controls-product-coverage`, `acceptance-product-observable`, `persistence-policy-explicit`, and `workflow-required`.
- Decision: Define one shared `AGENT_SETUP_PROMPT` that asks the coding agent to install from GitHub, open local Telegram authentication, present the real QR, verify with a read-only check, and ask before sends or paid/destructive actions. Render it through one reusable prompt variant in hero and final CTA.
- Alternatives rejected: Keeping npm first would preserve developer friction; a fake QR would misrepresent the live authentication flow; a short vague prompt could let agents skip safety or session verification; separate hero/final prompt strings could drift.
- State/output mapping: `AGENT_SETUP_PROMPT` feeds both copy buttons and clipboard state; prompt variant controls visible label and accessible copy-success names; hero DOM order makes prompt copy the first interactive element; `INSTALL_COMMAND` feeds only the final manual fallback.
- Files changed: content and tests, reusable copy component, hero, final CTA, responsive styles, Playwright acceptance, deterministic port-test fixture, landing spec/plan, and this worklog.
- Verification: `npm run verify:final` passed with 12 Node tests, 10 Vitest tests, production build, and 5 Playwright scenarios. Agent-controlled browser confirmed the prompt is the first hero interaction, QR copy is visible, npm is absent from hero, final fallback is present, overflow is zero, and console warnings/errors are zero. The port helper test also passed independently after its edge-port fixture was bounded to safe test ranges.
- Skipped checks: Full performance suite is not required for a Tier 2 copy/conversion change; ASCII renderer performance was unchanged after the already-verified detail pass.
- Risks: Coding agents can interpret setup prompts differently; the prompt pins the repository and safety boundaries but intentionally avoids shell-specific commands. Users still have GitHub quick start and manual npm fallback.

### Iteration 6 — Static outside-in testing proof

- Request: Replace the interactive “Test the experience, not just the endpoint” demo with useful static text.
- Task type: Tier 2 content hierarchy and responsive layout change.
- User-visible result: The section now contrasts what backend tests can prove with what a real Telegram user session can verify, then names the receipts available for follow-up debugging.
- Source/reference checked: The previous Scenario Lab content, the current landing narrative, and the live desktop/mobile section at `http://127.0.0.1:3002/#testing`.
- Reference inputs: Existing product claims and live page only; no new external asset.
- Docs/contracts read: Toolcraft brainstorming, writing-plans, workflow, existing landing spec/plan, and browser workflow.
- Contract rules applied: `controls-product-coverage`, `acceptance-product-observable`, and `workflow-required`; timeline and layers are not needed for a static evidence comparison.
- Decision: Use a typed two-sided comparison: backend evidence on the left, real-session evidence on the right, a restrained “real chat” bridge, one takeaway, and four concrete artifact labels.
- Alternatives rejected: Another fake transcript would look like a staged demo; a prompt lab would preserve the interaction the user explicitly removed; general capability cards would duplicate the use-case section without explaining the testing gap.
- State/output mapping: `OUTSIDE_IN_COMPARISON` feeds semantic, static DOM. The section has no local state, tabs, buttons, autoplay, or timeline.
- Files changed: Typed landing content and tests, a new outside-in component, route composition, responsive styles, browser acceptance, feature spec/plan, and this worklog.
- Verification: Typecheck passed; six focused content tests and all five landing Playwright scenarios passed. Agent-controlled browser review confirmed the two-column desktop and stacked mobile compositions, zero section controls, zero horizontal overflow, and no console warnings or errors.
- Skipped checks: The full Toolcraft performance suite is unnecessary for a static Tier 2 DOM/CSS change; quick verification, production build, responsive browser acceptance, and visual inspection cover the changed path.
- Risks: The old unmounted Scenario Lab source and styles remain available for a future product demo but are no longer part of the public route.

### Iteration 7 — Cloudflare Pages production delivery

- Request: Commit and push the completed landing, publish it on Cloudflare, and connect `agent-telegram.com`.
- User-visible result: The production build is deployed as Workers Static Assets and served directly from `https://agent-telegram.com/`; `https://telegram-agent.pages.dev/` remains a deployment fallback.
- Deployment mapping: `wrangler.jsonc` is the source of truth for the static asset directory, SPA fallback, compatibility date, and the `agent-telegram.com` Custom Domain. `npm run deploy:cloudflare` builds and deploys that configuration.
- Verification: Cloudflare returned a successful Workers deployment, attached the custom domain to `agent-telegram-site`, issued a certificate, and created DNS on the active zone. The domain and Pages fallback return the expected document title; `npm run verify:quick` and the production build pass.
- External status: The correct Cloudflare zone is active on `lou.ns.cloudflare.com` and `robin.ns.cloudflare.com`; the mistaken `telegram-agent.com` Pages association was removed.

### Iteration 8 — Frameless hero and native push deployment

- Request: Remove the screen/application framing around the hero duck and update production automatically on every push to `main`.
- Task type: Tier 2 hero composition change plus production delivery configuration.
- User-visible result: The ASCII duck now lives directly in the hero with only a soft amber glow; the window border, toolbar, dots, technical labels, grid, scanlines, footer, and badge are gone.
- Renderer decision: Keep the existing Canvas renderer, source video, responsive density, 12 fps cap, offscreen pause, and reduced-motion still unchanged. Only the presentation wrapper changes.
- Deployment decision: Native Cloudflare Workers Builds is connected to `dukaev/agent-telegram-site`, rooted at `/agent-telegram-site`, listening only to `main`, building with `npm run build`, and deploying with `npx wrangler deploy`. Non-production builds are disabled, and no GitHub deployment secret is used.
- Verification: The new acceptance test first failed on the missing frameless stage, then passed after implementation. `npm run verify:quick`, `npm run build`, and all five landing Playwright scenarios pass. Controlled-browser review at 1280 px and 390 px confirms no frame chrome and zero horizontal overflow. Cloudflare displays the saved Git repository, build command, deploy command, root directory, production branch, and disabled non-production builds.
- Recovery: Manual `npm run deploy:cloudflare` remains available if the native build service is unavailable.

### Iteration 9 — CLI and skill installation onboarding

- Request: Replace the QR-auth setup prompt with explicit npm and skill-install commands, update the final CTA to match, and push the change.
- Task type: Tier 0 copy-only conversion update.
- User-visible result: The shared agent prompt and final CTA now install the global CLI with `npm install -g agent-telegram`, add the agent skill with `agent-telegram skills install agent-telegram`, and retain the read-only verification and safety-confirmation language. The final CTA no longer mentions QR authentication.
- Source/reference checked: Existing shared landing content, final CTA component, content tests, and live desktop/mobile renders at `http://127.0.0.1:3002/#install`.
- Docs/contracts read: `AGENTS.md`, Toolcraft workflow, GitHub publish workflow, and browser workflow.
- Contract rules applied: `workflow-required`; no schema, runtime, renderer, state, export, timeline, or layer behavior changed.
- Decision: Define the skill command next to the npm command and reuse both constants in the prompt, manual fallback, and tests so the two installation paths cannot drift.
- Alternatives rejected: Retaining QR copy would contradict the requested flow; embedding a second unshared command string in the CTA would allow future content drift.
- State/output mapping: `INSTALL_COMMAND` and `SKILL_INSTALL_COMMAND` feed the shared `AGENT_SETUP_PROMPT` and final manual setup line; `AGENT_SETUP_PROMPT` remains the source for both copy buttons.
- Files changed: landing content, final CTA, focused content tests, and this worklog.
- Verification: `npm run typecheck` and six focused Vitest content tests passed. Controlled-browser review at desktop and mobile widths confirmed the updated heading, both steps, wrapped command copy, zero page overflow, and no console warnings or errors.
- Skipped checks: Full browser acceptance and performance suites are not required for a Tier 0 copy-only pass; focused responsive visual verification covers the only layout risk.
- Risks: Agent runtimes may format or execute the two commands differently, but the prompt preserves read-only verification and explicit confirmation boundaries.

### Iteration 10 — Four-second ping-pong hero loop

- Request: Make the hero duck play for two seconds, reverse the same motion for two seconds, and repeat continuously without changing the page interaction.
- Task type: Tier 3 animation/media output refinement; the renderer workload and application behavior are unchanged.
- User-visible result: The duck now completes one two-second action and naturally retraces it for two seconds, producing a four-second ping-pong loop without a hard reset.
- Source/reference checked: The existing `public/duck-laptop.mp4` was inspected with `ffprobe` and sampled at source times 0.0, 0.5, 1.0, 1.5, and approximately 2.0 seconds. The output reference contact sheet at `/tmp/duck-pingpong-contact.png` covers frames 0, 12, 24, 36, 47, 48, 59, 72, 83, and 95.
- Reference inputs: `public/duck-laptop.mp4`, `/tmp/duck-laptop-pingpong.mp4`, and `/tmp/duck-pingpong-contact.png`.
- Docs/contracts read: `AGENTS.md`, Toolcraft workflow, timeline animation, performance, decision contract, component rules, acceptance testing, reference study, brainstorming, writing plans, executing plans, and browser control.
- Contract rules applied: `video-reference-analysis`, `timeline-mode-choice`, `acceptance-product-observable`, `performance-coverage-levels`, and `workflow-required`.
- Decision: Bake the first 48 decoded source frames followed by those frames in reverse into the existing MP4 path. Keep the native forward-playing `<video loop>` and the Canvas ASCII renderer unchanged, with no visible transport or timeline because the mascot remains autonomous decorative media.
- Alternatives rejected: Negative `playbackRate` is not a reliable cross-browser reverse-video contract. Repeated `currentTime` backward seeks would make the browser decode against a source with only an initial keyframe and could introduce visible stutter.
- State/output mapping: Native video time advances from 0 to 4 seconds; encoded frames 0–47 show the forward action and frames 48–95 retrace it. Canvas continues sampling at its 12 fps cap, offscreen visibility still controls playback, and reduced motion still pauses on the visually unchanged opening frame.
- Files changed: `public/duck-laptop.mp4`, the ping-pong design spec and implementation plan, this worklog, and one stale reduced-motion heading expectation introduced by the concurrently merged onboarding copy. No React, Canvas, CSS, autoplay, visibility, reduced-motion, accessibility, or deployment code changed.
- Verification: `ffprobe` reports H.264, `yuv420p`, 640×688, 24 fps, 4.000 seconds, 96 frames, and no audio. Mirrored frame pairs scored SSIM 0.996708 and 0.997890; the opening frame scored 0.998908 against the previous asset. `npm run verify:quick`, `npm run build`, and all five `e2e/landing.spec.ts` Playwright scenarios pass. A controlled browser observed the 4-second duration, an active 129×72 ASCII grid, the loop boundary, zero horizontal overflow, and no warnings or errors.
- Skipped checks: The full Toolcraft performance suite is unnecessary because the Canvas dimensions, 12 fps cap, sample density, scheduling, visibility pause, and media dimensions did not change.
- Risks: The baked reverse half adds one generation of H.264 compression and may repeat an endpoint for one 24 fps frame. CRF 18 and paired-frame inspection keep the visual difference below a perceptible threshold for the ASCII renderer.

## Product State Mapping

- Testing proof: static backend-versus-real-session evidence, takeaway, and receipt labels; no demo state or transport controls.
- Navigation state: mobile menu visibility and anchor navigation.
- Conversion state: clipboard success feedback for the shared agent setup prompt; npm and skill installation are repeated as the non-primary manual fallback.
- Visual state: section reveal, use-case card tilt, hero timeline, architecture signal progress, and pointer spotlight.
- Content source: `src/landing/content.ts` centralizes scenarios, proof points, use cases, safety facts, and architecture steps.

## Implementation

- Renderer: Semantic React DOM and responsive CSS for the page, plus a bounded Canvas 2D ASCII pass for the hero video. Canvas is isolated to the media effect; semantic product content remains DOM.
- Controls: Prompt-copy actions, mobile menu, anchored navigation, and GitHub links. The outside-in testing section is intentionally static.
- Animation: Transform/opacity-first transitions, short hover feedback, spring state changes, one scroll-linked GSAP sequence, and one autonomous four-second ping-pong hero video loop. The ASCII loop has no visible transport or timeline.
- Performance: Pointer and tilt effects update through `requestAnimationFrame` and direct style transforms rather than React render loops. ASCII video drawing is capped at 12 fps, samples at most 144×80 cells, pauses offscreen, and never writes frame state through React. Mobile and reduced-motion modes disable nonessential continuous effects.
- Accessibility: Semantic landmarks, keyboard-capable buttons and links, visible focus treatment, descriptive labels, live clipboard feedback, and a complete `prefers-reduced-motion` mode.
- SEO: Product-specific title, description, theme color, canonical social description, and Open Graph metadata.

## Verification

- Passed: `npm run ai:check`.
- Passed: `npm run typecheck`.
- Passed: `npm run test` (repository integrity plus landing content tests).
- Passed: `npm run build`.
- Passed after the prompt-first pass: `npm run verify:final` (12 Node, 10 Vitest, 5 Playwright).
- Browser acceptance covers the populated/advancing ASCII canvas, removal of the old hero mockup, primary CTA, static outside-in evidence, absence of testing-section controls, copy feedback, mobile overflow/navigation, and reduced-motion output.
- Browser performance checkpoint: Passed with an agent-controlled browser at desktop width; no horizontal overflow and no console warnings or errors were observed. Mobile overflow is also asserted at 390 px.
- Visual inspection: The refreshed portrait line-art duck was reviewed in the hero after the GSAP entrance settled; its white background is removed, its silhouette remains readable in warm ASCII glyphs, and the frame reports the updated 4.04-second loop.

## Risks and Follow-ups

- The production bundle is approximately 602 kB uncompressed (about 196 kB gzip) and triggers Vite's advisory chunk-size warning; this is acceptable for the first interactive pass but can be reduced with route/vendor splitting if the site expands.
- GitHub popularity is described qualitatively rather than fetched live, avoiding a stale hard-coded star count and an extra request.
- Clipboard feedback depends on browser clipboard support; the install command always remains visible and selectable as a fallback.
