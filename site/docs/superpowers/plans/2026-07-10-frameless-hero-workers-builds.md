# Frameless Hero and Workers Builds Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the application-window chrome around the hero duck and make pushes to `main` deploy `agent-telegram.com` through native Cloudflare Workers Builds.

**Architecture:** Keep the existing Canvas renderer and animation lifecycle, but mount it in a presentation-only hero stage with an ambient glow and no interface chrome. Keep `wrangler.jsonc` as the deployment source of truth and connect the existing Worker directly to the private GitHub repository through Cloudflare Workers Builds.

**Tech Stack:** React 19, TypeScript, Vite 8, Canvas 2D, GSAP, Playwright, Cloudflare Workers Static Assets, Wrangler 4, Cloudflare Workers Builds.

## Global Constraints

- The hero must not display a frame, toolbar, dots, title, FPS label, source label, footer, badge, grid, or scanline overlay.
- Preserve the current ASCII renderer, source video, frame density, 12 fps cap, autonomous loop, offscreen pause, and reduced-motion still frame.
- Preserve the prompt-first onboarding copy and existing hero entrance choreography.
- Deploy only pushes to `main`; non-production branch builds remain disabled.
- The Worker name remains `agent-telegram-site`, and `wrangler.jsonc` remains the deployment source of truth.
- Do not add GitHub Actions or store a short-lived Wrangler OAuth token in GitHub.

---

### Task 1: Replace Hero Window Chrome with a Frameless Duck Stage

**Files:**
- Modify: `e2e/landing.spec.ts`
- Modify: `src/landing/components/hero.tsx`
- Modify: `src/styles.css`

**Interfaces:**
- Consumes: `AsciiDuckVideo(): React.JSX.Element`, `[data-demo-reveal]`, `.ascii-duck`, and `.ascii-duck__canvas`.
- Produces: `.hero__duck-stage`, a presentation-only wrapper containing the unchanged ASCII renderer.

- [ ] **Step 1: Add failing browser assertions for frameless output**

In the primary landing test, keep the existing canvas visibility and frame-progress assertions, then add:

```ts
await expect(page.locator(".hero__duck-stage")).toBeVisible();
await expect(page.locator(".ascii-frame")).toHaveCount(0);
await expect(page.getByText("DUCK.ASCII")).toHaveCount(0);
await expect(page.getByText("12 FPS")).toHaveCount(0);
await expect(page.getByText("CANVAS 2D · LOCAL MEDIA")).toHaveCount(0);
```

- [ ] **Step 2: Run the targeted test and confirm the new contract fails**

Run: `npx playwright test e2e/landing.spec.ts --grep "landing tells the real-user Telegram story"`

Expected: FAIL because `.hero__duck-stage` does not exist and the frame labels are still rendered.

- [ ] **Step 3: Simplify the hero markup**

Replace the complete `.ascii-frame` subtree in `Hero` with:

```tsx
<div className="hero__ascii-wrap" data-demo-reveal>
  <div className="hero__duck-stage">
    <AsciiDuckVideo />
  </div>
</div>
```

Do not change the copy column, actions, `data-demo-reveal`, or `AsciiDuckVideo`.

- [ ] **Step 4: Replace frame CSS with the frameless stage**

Keep `.hero__ascii-wrap` as a positioned layout wrapper. Remove the `.ascii-frame`, `.ascii-frame::after`, `.ascii-frame__topbar`, `.ascii-frame__footer`, `.ascii-frame__title`, `.ascii-frame__rate`, `.ascii-frame__stage`, `.ascii-frame__grid`, `.ascii-frame__index`, and `.ascii-frame__badge` rules. Do not remove `.window-dots`, because the unmounted Telegram demo still owns that shared style.

Add this stage contract:

```css
.hero__duck-stage {
  position: relative;
  display: grid;
  min-height: 610px;
  place-items: center;
  isolation: isolate;
}

.hero__duck-stage::before {
  position: absolute;
  z-index: -1;
  width: 72%;
  aspect-ratio: 1;
  border-radius: 50%;
  background: rgba(238, 177, 53, 0.055);
  content: "";
  filter: blur(82px);
  pointer-events: none;
}
```

Keep `.ascii-duck` and `.ascii-duck__canvas`, increasing the duck only if the desktop composition needs it after browser inspection. At `max-width: 760px`, set `.hero__duck-stage { min-height: 390px; }` and retain `.ascii-duck { width: 88%; }`.

- [ ] **Step 5: Run the targeted browser test**

Run: `npx playwright test e2e/landing.spec.ts --grep "landing tells the real-user Telegram story"`

Expected: PASS with the advancing ASCII canvas and zero frame chrome.

---

### Task 2: Verify Responsive and Reduced-Motion Behavior

**Files:**
- Modify: `e2e/landing.spec.ts` only if an observable regression needs coverage
- Modify: `docs/toolcraft/agent-worklog.md`

**Interfaces:**
- Consumes: `.hero__duck-stage`, `[data-ascii-canvas]`, the existing mobile overflow assertion, and the reduced-motion frame-time observable.
- Produces: verified desktop/mobile/reduced-motion output and an implementation decision trail.

- [ ] **Step 1: Run all local verification**

Run: `npm run verify:quick`

Expected: 12 Node tests and 11 landing Vitest tests pass.

Run: `npm run build`

Expected: TypeScript and Vite production build pass; the existing bundle-size advisory may remain.

Run: `npx playwright test e2e/landing.spec.ts`

Expected: all five landing scenarios pass.

- [ ] **Step 2: Inspect desktop output in the controlled browser**

Open `http://127.0.0.1:3002/#top`, wait for the GSAP entrance, and confirm:

- the duck appears directly on the page;
- no visual border, toolbar, label, grid, or badge surrounds it;
- the glow supports the silhouette without reading as a surface;
- `document.documentElement.scrollWidth - document.documentElement.clientWidth` equals `0`;
- the console contains no warnings or errors.

- [ ] **Step 3: Inspect mobile output at 390 × 844**

Temporarily set the controlled browser viewport to `390 × 844`, inspect the hero, confirm zero horizontal overflow and a balanced gap between copy and duck, then reset the viewport.

- [ ] **Step 4: Record the completed composition change**

Append an iteration to `docs/toolcraft/agent-worklog.md` naming the removed chrome, preserved renderer behavior, responsive checks, test results, and the approved native Workers Builds decision.

---

### Task 3: Connect Native Workers Builds and Prove Push-to-Production

**Files:**
- Verify: `wrangler.jsonc`
- Verify: `package.json`
- No `.github/workflows` file is created

**Interfaces:**
- Consumes: Worker `agent-telegram-site`, repository `dukaev/agent-telegram-site`, `wrangler.jsonc`, and `npm run build`.
- Produces: one Cloudflare production build trigger for pushes to `main`, deploying the existing Custom Domain `agent-telegram.com`.

- [ ] **Step 1: Connect the existing Worker to GitHub**

In Cloudflare Workers & Pages, open `agent-telegram-site` → Settings → Builds → Connect. Authorize the Cloudflare Workers & Pages GitHub App for `dukaev/agent-telegram-site` if required, then select that repository.

Use exactly:

```text
Production branch: main
Root directory: /agent-telegram-site
Build command: npm run build
Deploy command: npx wrangler deploy
Non-production branch builds: disabled
```

Stop and request user authentication if the Cloudflare GitHub App authorization screen cannot be completed in the available signed-in browser. Do not create a temporary GitHub secret.

- [ ] **Step 2: Confirm the build configuration matches the repository**

Verify `wrangler.jsonc` contains `"name": "agent-telegram-site"`, `"directory": "./dist"`, and the `agent-telegram.com` Custom Domain. Verify `package.json` has a passing `build` script and the Wrangler dev dependency.

- [ ] **Step 3: Commit the implementation**

Stage the plan, hero, styles, acceptance, and worklog files. Commit with:

```text
remove hero chrome and enable Workers Builds
```

- [ ] **Step 4: Push `main` and capture the commit SHA**

Run: `git push origin main`

Run: `git rev-parse HEAD`

Expected: the local `main` commit is present on `origin/main`.

- [ ] **Step 5: Confirm Cloudflare reports the push build**

Query GitHub check runs for the pushed SHA:

```bash
gh api repos/dukaev/agent-telegram-site/commits/$(git rev-parse HEAD)/check-runs \
  --jq '.check_runs[] | {name, status, conclusion, details_url}'
```

Expected: a Cloudflare check run reaches `completed` with conclusion `success`. If the check is still queued or running, poll without making another deployment.

- [ ] **Step 6: Verify production output**

Run:

```bash
curl --silent --show-error --fail --location --max-time 25 \
  --output /tmp/agent-telegram-production.html \
  --write-out 'status=%{http_code} url=%{url_effective}\n' \
  https://agent-telegram.com/
```

Expected: HTTP 200. Confirm the returned HTML references the current production asset hashes, and inspect the live hero once to verify the frameless output.
