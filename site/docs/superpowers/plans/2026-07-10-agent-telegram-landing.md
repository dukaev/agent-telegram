# Agent Telegram Landing Implementation Plan

> **For agentic workers:** Execute inline in this session. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a production-ready, interactive React/Vite landing page for agent-telegram with Motion and GSAP.

**Architecture:** Render a dedicated marketing page from the index route, with copy isolated from interactive components. Motion owns user-driven transitions; GSAP owns rare explanatory and scroll-linked sequences. Preserve the copied Toolcraft runtime but remove it from the public route.

**Tech Stack:** React 19, Vite 8, TypeScript 6, Motion for React, GSAP, Phosphor Icons, Playwright.

## Global Constraints

- English copy only.
- Primary use case: real-user end-to-end Telegram bot testing.
- Do not claim native MCP support that is not present in the repository.
- Provide reduced-motion fallbacks for every spatial animation.
- Keep GitHub and install CTAs functional.
- Use transforms and opacity for continuous animation.

---

### Task 1: Landing foundation and content contract

**Files:**
- Create: `src/landing/content.ts`
- Create: `src/landing/use-copy-command.ts`
- Create: `src/landing/use-reduced-motion.ts`
- Modify: `package.json`

**Interfaces:**
- Produces `BOT_SCENARIOS`, `USE_CASES`, `SAFETY_POINTS`, `PROOF_POINTS`, and `INSTALL_COMMAND`.
- Produces `useCopyCommand(command)` returning `{ copied, copy }`.
- Produces `useReducedMotionPreference()` returning a boolean.

- [ ] Add `gsap` as a runtime dependency and use the existing `motion` package through `motion/react`.
- [ ] Define typed content arrays with stable IDs for interaction and browser assertions.
- [ ] Implement clipboard success state with a two-second reset and cleanup.
- [ ] Implement a reactive `matchMedia('(prefers-reduced-motion: reduce)')` hook.
- [ ] Run `npm run typecheck`; expect PASS.

### Task 2: Interactive navigation and hero

**Files:**
- Create: `src/landing/components/site-header.tsx`
- Create: `src/landing/components/copy-command.tsx`
- Create: `src/landing/components/telegram-demo.tsx`
- Create: `src/landing/components/hero.tsx`

**Interfaces:**
- `SiteHeader` renders anchor navigation and install/GitHub actions.
- `CopyCommand` consumes a command string and exposes copy feedback.
- `TelegramDemo` consumes an active scenario ID and exposes restart/play controls.
- `Hero` owns the GSAP entrance context and composes the demo.

- [ ] Build a sticky accessible header with mobile-safe navigation.
- [ ] Build the install command control with Copy/Check state morphing.
- [ ] Build the terminal/chat demo from typed scenario events.
- [ ] Add Motion shared-layout selection and spring cursor response.
- [ ] Add a GSAP hero entrance timeline with reduced-motion bypass.
- [ ] Run `npm run typecheck`; expect PASS.

### Task 3: Product story sections

**Files:**
- Create: `src/landing/components/proof-strip.tsx`
- Create: `src/landing/components/scenario-lab.tsx`
- Create: `src/landing/components/use-cases.tsx`
- Create: `src/landing/components/architecture.tsx`
- Create: `src/landing/components/safety.tsx`
- Create: `src/landing/components/final-cta.tsx`

**Interfaces:**
- `ScenarioLab` owns the selected scenario ID and renders command/output details.
- `UseCases` renders pointer-tilt cards without pointer-move React rerenders.
- `Architecture` owns a GSAP ScrollTrigger context and renders the signal path.
- Remaining sections are presentational and read typed content.

- [ ] Implement the proof strip and section reveal primitive.
- [ ] Implement selectable bot scenarios with live terminal and chat state.
- [ ] Implement four interactive use-case cards with fine-pointer tilt.
- [ ] Implement the architecture path and scroll-linked signal.
- [ ] Implement safety evidence and the final conversion section.
- [ ] Run `npm run typecheck`; expect PASS.

### Task 4: Page composition and visual system

**Files:**
- Modify: `src/routes/index.tsx`
- Replace: `src/styles.css`
- Modify: `index.html`

**Interfaces:**
- `AppHome` composes semantic page landmarks and global decorative layers.
- CSS class contracts match component markup and responsive breakpoints.

- [ ] Compose the route from focused landing components.
- [ ] Add the complete dark visual system, layout, responsive rules, focus styles, and reduced-motion rules.
- [ ] Add a requestAnimationFrame-backed pointer spotlight without React rerenders.
- [ ] Update title, description, theme color, Open Graph metadata, and canonical product language.
- [ ] Run `npm run build`; expect PASS.

### Task 5: Browser acceptance and delivery gate

**Files:**
- Create: `e2e/landing.spec.ts`
- Modify: `playwright.config.ts`
- Modify: `docs/toolcraft/agent-worklog.md`

**Interfaces:**
- Playwright discovers only landing acceptance specs for this public route.
- Tests verify the core product story, interactions, responsive layout, and reduced-motion behavior.

- [ ] Test hero content, GitHub link, navigation anchors, and install command.
- [ ] Test scenario selection and demo state change.
- [ ] Test clipboard feedback with granted browser permissions.
- [ ] Test mobile layout for horizontal overflow and CTA visibility.
- [ ] Test reduced-motion mode for content availability and disabled autoplay semantics.
- [ ] Record the deliberate public-route boundary and verification evidence in the worklog.
- [ ] Run `npm run ai:check`, `npm run test`, `npm run build`, and `npm run test:browser`; expect PASS.
- [ ] Start the dev server and perform desktop/mobile visual review.

### Task 6: Prompt-first onboarding

**Files:**
- Modify: `src/landing/content.ts`
- Modify: `src/landing/content.test.ts`
- Modify: `src/landing/components/copy-command.tsx`
- Modify: `src/landing/components/hero.tsx`
- Modify: `src/landing/components/final-cta.tsx`
- Modify: `src/styles.css`
- Modify: `e2e/landing.spec.ts`
- Modify: `docs/toolcraft/agent-worklog.md`

- Define one shared agent setup prompt with QR authentication, read-only verification, and explicit confirmation boundaries.
- Generalize the copy component with prompt-specific semantics and a multiline presentation.
- Put the prompt before hero navigation actions and follow it immediately with the one-QR explanation.
- Rewrite the final CTA as the same two-step flow and demote npm install to a manual fallback.
- Update content and browser acceptance for exact clipboard output, prompt-first ordering, QR copy, mobile layout, and prior interactions.
- Run `npm run verify:final`, inspect the live page, then commit all current workspace changes and push `main`.

### Task 7: Replace scenario lab with static outside-in explanation

**Files:**
- Create: `src/landing/components/outside-in.tsx`
- Modify: `src/landing/content.ts`
- Modify: `src/landing/content.test.ts`
- Modify: `src/routes/index.tsx`
- Modify: `src/styles.css`
- Modify: `e2e/landing.spec.ts`
- Modify: `docs/toolcraft/agent-worklog.md`

- Define a typed static comparison for backend evidence, real-user evidence, and run artifacts.
- Replace `ScenarioLab` in route composition while preserving the `#testing` anchor.
- Render two static evidence columns, one directional bridge, and a concise takeaway/artifact strip.
- Do not add tabs, controls, playback, local state, or fake chat output.
- Replace scenario interaction acceptance with static content, no-tabs, mobile, and reduced-motion assertions.
- Run `npm run verify:quick`, `npm run build`, landing Playwright, and agent-browser desktop/mobile visual checks.
