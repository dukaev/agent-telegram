# Landing Site Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import the current `dukaev/agent-telegram-site` application into `site/` while preserving its Cloudflare deployment behavior and keeping it isolated from the existing root Node/Vite application.

**Architecture:** `site/` is a self-contained Vite application with its own dependencies, tests, Toolcraft runtime, and Wrangler configuration. The repository root delegates development, build, test, and deployment commands to that directory through npm's `--prefix` option; no workspace or dependency merge is introduced.

**Tech Stack:** Go CLI, npm, React 19, TypeScript 6, Vite 8, Playwright, Vitest, Wrangler 4, Cloudflare Workers static assets.

## Global Constraints

- Import source commit `0cc31f9fa5f270e6aecbd6b457ff0a9fe890fc23` without its Git history.
- Keep the landing app under `site/` with its own `package.json` and `package-lock.json`.
- Preserve `agent-telegram.com`, compatibility date `2026-07-10`, `./dist`, and SPA fallback behavior in `site/wrangler.jsonc`.
- Do not merge the landing app with `web/auth` or convert the repository to npm workspaces.
- Do not redesign or change landing-page behavior.
- Do not publish a live Cloudflare deployment.
- Verification tier: Tier 4, because a complete application, dependency tree, tests, deployment configuration, and runtime documentation are imported.

---

### Task 1: Import the landing application snapshot

**Files:**
- Create: `site/` from the nested `agent-telegram-site/` directory at source commit `0cc31f9fa5f270e6aecbd6b457ff0a9fe890fc23`
- Exclude: source `.git/`, outer `img.jpg`, outer `vid.mp4`, and the outer wrapper `.gitignore`
- Verify: `site/package.json`, `site/package-lock.json`, `site/vite.config.ts`, `site/wrangler.jsonc`, `site/src/`, `site/public/`, `site/e2e/`, `site/scripts/`, `site/docs/`, `site/LICENSE.md`, `site/NOTICE.md`

**Interfaces:**
- Consumes: Git snapshot `0cc31f9fa5f270e6aecbd6b457ff0a9fe890fc23` from `https://github.com/dukaev/agent-telegram-site.git`.
- Produces: a self-contained npm project at repository-relative path `site/`.

- [ ] **Step 1: Fetch the exact source snapshot into a temporary directory**

```bash
tmp_dir="$(mktemp -d)"
git clone --quiet https://github.com/dukaev/agent-telegram-site.git "$tmp_dir/source"
git -C "$tmp_dir/source" checkout --quiet 0cc31f9fa5f270e6aecbd6b457ff0a9fe890fc23
```

Expected: `$tmp_dir/source/agent-telegram-site/package.json` exists and `git -C "$tmp_dir/source" rev-parse HEAD` prints `0cc31f9fa5f270e6aecbd6b457ff0a9fe890fc23`.

- [ ] **Step 2: Import only the nested application directory**

Apply the snapshot of `$tmp_dir/source/agent-telegram-site/` at repository path `site/`, preserving executable bits and excluding all Git metadata. Do not import the source repository's outer `img.jpg`, `vid.mp4`, or `.gitignore`.

- [ ] **Step 3: Verify snapshot completeness**

```bash
diff -ru --exclude node_modules --exclude dist "$tmp_dir/source/agent-telegram-site" site
```

Expected: no output and exit status 0.

- [ ] **Step 4: Record the imported snapshot**

```bash
git add site
git commit -m "Import landing site application"
```

Expected: the commit contains only the new `site/` tree.

### Task 2: Add root commands and deployment documentation

**Files:**
- Modify: `package.json`
- Modify: `README.md`

**Interfaces:**
- Consumes: the npm scripts `dev`, `build`, `test`, and `deploy:cloudflare` from `site/package.json`.
- Produces: root scripts `site:dev`, `site:build`, `site:test`, and `site:deploy`.

- [ ] **Step 1: Add root npm delegation scripts**

Add these exact entries to the root `scripts` object without changing existing auth-web scripts:

```json
"site:dev": "npm --prefix site run dev",
"site:build": "npm --prefix site run build",
"site:test": "npm --prefix site run test",
"site:deploy": "npm --prefix site run deploy:cloudflare"
```

- [ ] **Step 2: Document landing development and Cloudflare configuration**

Append this section under `## Development` in the root README, before the existing `DEVELOPMENT.md` sentence:

````markdown
### Landing site

The public landing page is a self-contained application in `site/`.

```bash
npm --prefix site install
npm run site:dev
npm run site:test
npm run site:build
```

Deploy it from the repository root with `npm run site:deploy`. For Cloudflare
Git integration, set the project root directory to `site`; the Wrangler
configuration in that directory remains the deployment source of truth for
`agent-telegram.com`.
````

- [ ] **Step 3: Verify root script delegation metadata**

```bash
node -e 'const p=require("./package.json"); for (const key of ["site:dev","site:build","site:test","site:deploy"]) { if (!p.scripts[key]) throw new Error(`missing ${key}`) }'
```

Expected: exit status 0 with no output.

- [ ] **Step 4: Commit root integration**

```bash
git add package.json README.md
git commit -m "Integrate landing site workflows"
```

Expected: the commit contains only `package.json` and `README.md`.

### Task 3: Verify the imported application and Cloudflare configuration

**Files:**
- Generated and ignored: `site/node_modules/`, `site/dist/`, `site/.toolcraft-port.json` if created by the site scripts
- Inspect: `site/docs/toolcraft/agent-worklog.md`
- Modify only if verification evidence is required by the imported contract: `site/docs/toolcraft/agent-worklog.md`

**Interfaces:**
- Consumes: root delegation scripts from Task 2 and the complete `site/` application from Task 1.
- Produces: passing landing tests/build, a validated Wrangler dry run, a browser-verified local landing page, and a passing existing auth-web check.

- [ ] **Step 1: Install exact landing dependencies**

```bash
npm --prefix site ci
```

Expected: dependencies install from `site/package-lock.json` without changing the lockfile.

- [ ] **Step 2: Run the landing final gate**

```bash
npm --prefix site run verify:final
```

Expected: AI preflight, docs checks, integrity checks, unit tests, TypeScript build, Vite production build, and non-performance Playwright tests all pass.

- [ ] **Step 3: Run the required first-working-version performance gate**

Use the controlled browser when available. If it cannot execute the app's performance scenarios, run:

```bash
npm --prefix site run verify:perf
```

Expected: all `browser perf:` Playwright scenarios pass sequentially.

- [ ] **Step 4: Validate Cloudflare packaging without deployment**

```bash
npm --prefix site exec -- wrangler deploy --dry-run
```

Expected: Wrangler resolves `site/wrangler.jsonc`, packages `site/dist`, reports the `agent-telegram.com` custom domain configuration, and does not deploy.

- [ ] **Step 5: Verify the existing auth web application**

```bash
npm run check:web:auth
```

Expected: TypeScript and Vite auth-web checks pass.

- [ ] **Step 6: Start and inspect the landing page**

```bash
npm run site:dev
```

Expected: the site script reports its verified local URL. Open that URL, confirm the hero and primary interaction render, check the browser console for errors, and exercise the main landing-page flow covered by `site/e2e/landing.spec.ts`.

- [ ] **Step 7: Record verification evidence if the Toolcraft worklog requires an import entry**

Add a Decision Trail entry stating that this pass relocated the unchanged product into `site/`, preserved the schema/rendering/export behavior and Cloudflare configuration, ran Tier 4 verification, and introduced no product state/output mapping changes. Do not claim checks that did not pass.

- [ ] **Step 8: Commit any required verification record**

```bash
git add site/docs/toolcraft/agent-worklog.md
git commit -m "Record landing integration verification"
```

Expected: create this commit only when the worklog changed; otherwise leave the working tree unchanged.
