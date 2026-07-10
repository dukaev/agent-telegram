# Agent Telegram Landing Design

## Objective

Build a high-interaction, English-language landing page for `agent-telegram`: a local-first Telegram automation CLI for AI agents. The primary story is end-to-end Telegram bot testing from a real user session. Secondary stories are support workflows, community monitoring, and structured automation.

## Product Truth

- The product is currently a CLI, local IPC daemon, bundled agent skill, and optional HTTP/OpenAPI surface.
- The page must not claim that the repository already ships a native MCP server.
- The core differentiator is real-user Telegram access over MTProto rather than the Bot API.
- Authentication is local, destructive and paid operations require explicit confirmation, and macOS can persist sessions in Keychain.

## Audience And Conversion

The primary audience is bot developers, QA engineers, and AI-agent builders. The primary conversion is copying a complete setup prompt into an AI coding agent, followed by one Telegram QR scan. Secondary conversions are visiting GitHub and using the manual npm fallback.

## Information Architecture

1. Sticky glass navigation with section anchors, GitHub, and an install CTA.
2. Hero with product positioning, install command, proof chips, and an interactive terminal-to-Telegram bot flow.
3. Proof strip describing real-user sessions, structured JSON, safety contracts, and observability.
4. Bot-testing story with selectable scenarios and live command/chat output.
5. Use-case cards for QA, support, community monitoring, and workflow automation.
6. Architecture section showing agent to CLI/HTTP to local daemon to Telegram MTProto.
7. Safety section with local auth, Keychain, confirmation gates, and redacted logs.
8. Final install CTA and compact footer.

## Visual Direction

Use a dark, technical, editorial surface: near-black blue background, Telegram cyan, electric violet, and acid-lime status accents. Typography uses Inter Variable for UI and a system monospace for commands. The main visual motif is a live signal moving between terminal events and a Telegram conversation, supported by a faint grid and restrained glow.

## Interaction Model

- Motion for React handles interruptible UI state: scenario tabs, card entrances, button feedback, magnetic pointer response, and shared-layout selection.
- GSAP handles rare explanatory sequences: hero timeline, scroll-linked architecture signal, and staggered section reveals.
- The hero demo can be played, paused, restarted, or advanced by choosing a bot-testing scenario.
- Use-case cards tilt subtly toward the pointer on fine-pointer devices and expose a visible keyboard focus state.
- The install command copies to the clipboard and morphs to a success state.
- Pointer spotlight and parallax are decorative and disabled on touch or reduced-motion environments.
- `prefers-reduced-motion` keeps opacity and color feedback while disabling spatial movement, autoplay, parallax, and scroll scrubbing.

## Architecture

The requested artifact is a public landing page, not a Toolcraft editing product. The route therefore renders a dedicated landing tree directly. Existing Toolcraft runtime source remains untouched and unused so the landing can be a normal React/Vite site without editor chrome.

Focused modules:

- `src/routes/index.tsx`: page composition only.
- `src/landing/content.ts`: typed copy and scenario data.
- `src/landing/use-copy-command.ts`: clipboard behavior and transient status.
- `src/landing/use-reduced-motion.ts`: motion preference contract shared with GSAP.
- `src/landing/components/*`: navigation, hero demo, scenario lab, use cases, architecture, safety, and CTA.
- `src/styles.css`: visual system, responsive layout, and CSS-only interaction states.

## Accessibility And Performance

- Semantic landmarks, heading order, anchor navigation, visible focus, and descriptive button names are required.
- All core content remains readable and actionable with JavaScript animation disabled.
- Animations primarily use transform and opacity. Repeated hover interactions stay below 220ms.
- GSAP contexts are reverted on unmount. Pointer listeners use animation frames rather than React state on every move.
- Mobile removes card tilt and simplifies the hero composition to a single-column story.

## Verification

Verification tier: Tier 4 — the visible product and route assembly are replaced.

Run:

- `npm run ai:check`
- `npm run test`
- `npm run build`
- landing-focused Playwright browser acceptance
- browser visual review at desktop and mobile widths
- reduced-motion browser check

The legacy Toolcraft performance suite is not applicable because the delivered product is a marketing site rather than an editor/canvas application. Landing browser acceptance replaces the legacy canvas acceptance entry point.

## Prompt-First Integration Pass

Verification tier: Tier 2 — primary conversion behavior and copy mapping change; renderer, media, timeline, layers, persistence, and export remain unchanged.

- Primary conversion becomes copying a complete setup prompt into an AI coding agent.
- The prompt tells the agent to install `agent-telegram`, start local Telegram authentication, present the real QR code, verify the resulting session with a read-only check, and ask before sends, paid actions, or destructive actions.
- The user-facing integration story is exactly two steps: paste the prompt, then scan one Telegram QR code.
- `npm install -g agent-telegram` remains visible only as a manual fallback, not the first action.
- Hero and final CTA use the same prompt source so clipboard behavior cannot drift.
- The page must not display a fake QR code; the QR is produced by the real local authentication flow.
- Browser acceptance verifies prompt-first ordering, exact clipboard text, QR explanation, manual fallback, mobile layout, and the existing product interactions.

## Static Outside-In Testing Pass

Verification tier: Tier 2 — section content, route composition, and acceptance mapping change; renderer, media, timeline, layers, persistence, and export remain unchanged.

- Remove the interactive Acme Bot scenario lab and all fake conversation output from the public route.
- Keep `#testing` as the navigation target, but turn it into a static explanation of outside-in testing.
- Lead with the core failure mode: an endpoint or handler can pass while the Telegram conversation remains broken.
- Compare two complementary evidence sets:
  - backend/API tests prove handler execution, schema validity, Telegram acceptance, and callback success;
  - `agent-telegram` verifies chat placement, visible/usable buttons, ordered state transitions, and the complete user flow.
- Close with the explicit positioning: use API tests for implementation and a real-user session for experience.
- Name the useful output artifacts: message IDs, timings, actions, and structured receipts.
- No controls, timeline, layers, persistence, export, autoplay, tabs, or hidden interaction state belong to this section.
