# Frameless Hero and Workers Builds Design

## Goal

Remove the application-window metaphor from the hero mascot and make every push to `main` automatically rebuild and deploy the production site at `https://agent-telegram.com/`.

## Hero Composition

The ASCII duck remains the hero's only visual object. Remove the outer frame, window dots, title, FPS label, source index, grid, scanline overlay, footer, and Canvas badge. The visual must no longer resemble a program, terminal, player, or embedded screen.

Keep `AsciiDuckVideo` and its current Canvas renderer unchanged. Place it in a simple presentation wrapper that provides layout bounds and one soft, low-opacity amber glow behind the duck. The wrapper has no visible border, surface fill, toolbar, captions, or technical metadata.

The existing hero copy, prompt-first onboarding, QR cue, actions, GSAP entrance, autonomous loop, offscreen pausing, and reduced-motion still frame remain unchanged. The existing `data-demo-reveal` hook stays on the visual wrapper so entrance choreography does not drift.

## Responsive Behavior

- Desktop: the duck occupies the current right-hand hero column and reads as part of the page, not as content inside a component.
- Tablet: the duck remains centered below the copy and uses the existing maximum width constraint.
- Mobile: the visual stays within the viewport, with a slightly larger relative duck width and a shorter minimum presentation height.
- The glow may scale with the duck but must not create clipping or horizontal overflow.

## Automated Production Deployment

Use Cloudflare Workers Builds, not GitHub Actions. Connect the existing Worker `agent-telegram-site` to the private GitHub repository `dukaev/agent-telegram-site` through the Cloudflare Workers & Pages GitHub App.

Production configuration:

- Repository: `dukaev/agent-telegram-site`
- Root directory: `/agent-telegram-site`
- Production branch: `main`
- Build command: `npm run build`
- Deploy command: `npx wrangler deploy`
- Non-production branch builds: disabled
- Worker name: `agent-telegram-site`, matching `wrangler.jsonc`
- Production hostname: `agent-telegram.com`, retained through the existing Custom Domain configuration

The native integration owns the push trigger and Cloudflare build token. No expiring Wrangler OAuth token, Cloudflare API token, or deployment secret is stored in GitHub. The existing `npm run deploy:cloudflare` command remains available as a manual recovery path.

If the Cloudflare GitHub App has not yet been authorized for the repository, that one-time authorization is required before the repository connection can be created. Do not substitute the short-lived local Wrangler OAuth token in CI.

## Repository Changes

- Simplify `src/landing/components/hero.tsx` to render the duck without frame chrome.
- Replace frame-specific hero CSS in `src/styles.css` with frameless layout and glow styles; do not alter unrelated Telegram demo styles.
- Update browser acceptance to assert that the ASCII canvas remains visible and advancing while frame labels and chrome are absent.
- Update the Toolcraft worklog with the composition and deployment decisions.
- Do not add a `.github/workflows` deployment file because Cloudflare Workers Builds is the selected CI/CD owner.

## Verification

1. Run `npm run verify:quick` and `npm run build`.
2. Run the landing Playwright suite, including desktop, mobile, and reduced-motion coverage.
3. Inspect the hero in an agent-controlled browser at desktop and mobile widths, confirming no frame chrome and no horizontal overflow.
4. Connect Workers Builds with the exact production configuration above.
5. Push the implementation commit to `main` and confirm that Cloudflare creates a successful build for that commit.
6. Confirm `https://agent-telegram.com/` returns HTTP 200 and contains the new frameless hero output.

## Failure and Recovery

- A missing GitHub App authorization blocks only the repository connection; local checks and manual deployment remain usable.
- A failed Workers Build must not promote a partial deployment. The previous Worker version remains active until `wrangler deploy` succeeds.
- Manual recovery uses `npm run deploy:cloudflare` from the project root.

## Acceptance Criteria

- The first slide contains no visible frame, toolbar, dots, title, FPS label, source label, footer, badge, or grid around the duck.
- The ASCII duck remains animated for normal-motion users and static for reduced-motion users.
- The hero remains balanced on desktop and mobile with zero horizontal overflow.
- Every push to `main` triggers one production Workers Build for `agent-telegram-site`.
- Non-production branches do not deploy.
- A successful build updates `agent-telegram.com` without manual Wrangler execution.
