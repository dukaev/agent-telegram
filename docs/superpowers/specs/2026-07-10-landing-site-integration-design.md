# Landing Site Integration Design

## Goal

Move the current contents of `dukaev/agent-telegram-site` into the
`dukaev/agent-telegram` repository so the CLI and its public landing page are
maintained together. The source repository's Git history does not need to be
preserved. Existing Cloudflare deployment behavior for `agent-telegram.com`
must continue to work.

## Repository Structure

The landing page will live in a self-contained `site/` directory. It will keep
its own `package.json`, `package-lock.json`, Vite configuration, Wrangler
configuration, TypeScript configuration, source files, public assets, tests,
Toolcraft runtime, local documentation, and licensing notices.

The source repository's `.git` directory and outer wrapper files that are not
part of the nested application will not be copied. The existing CLI repository
files and its `web/auth` application will remain in place.

## Root Integration

The root `package.json` will expose convenience scripts that delegate to the
landing project without merging dependency trees:

- `site:dev`
- `site:build`
- `site:test`
- `site:deploy`

Each command will run the corresponding npm script with `site/` as its working
directory. The root and landing projects will retain separate lockfiles. This
avoids dependency and Vite configuration conflicts with the existing auth web
application.

The root README will document how to install landing dependencies, run the
development server, build and test the site, and deploy it.

## Cloudflare Deployment

The existing `site/wrangler.jsonc` will remain the source of truth for the
Cloudflare Worker/assets deployment and the `agent-telegram.com` custom domain.
Its asset directory remains `./dist`, resolved relative to `site/`.

Local or CI deployment from the repository root will use `npm run site:deploy`.
For Cloudflare Git integration, the project root directory must be configured
as `site`, with the existing site build/deploy commands executed from there.
No domain, compatibility date, or SPA fallback behavior will be changed during
the move.

## Verification

This is a Tier 4 change because it imports a complete application, dependencies,
tests, deployment configuration, and local runtime documentation.

After copying the project:

1. Install the landing dependencies from its lockfile.
2. Run the landing project's AI preflight and final verification suite.
3. Run its production build and confirm `site/dist` is generated.
4. Validate the Wrangler deployment configuration without publishing a live
   deployment.
5. Start the development server and verify the landing page in a browser.
6. Confirm the existing root auth-web check still passes.

Live deployment is outside this change: publishing would modify external state.
The repository will be left ready for the existing Cloudflare project to build
from `site/`.

## Non-goals

- Preserve the source repository's Git history.
- Merge the landing and auth web applications.
- Convert the repository to npm workspaces.
- Redesign or otherwise change the landing page.
- Change Cloudflare domains or publish a production deployment.
