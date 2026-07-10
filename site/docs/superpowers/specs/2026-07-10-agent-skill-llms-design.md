# Agent Skill And `llms.txt` Design

## Objective

Publish the canonical `agent-telegram` Codex skill and full LLM-facing CLI reference on the product website, then make both resources discoverable from the landing page.

## Source Of Truth

Use `dukaev/agent-telegram` `main` as the only product-information source.

- Copy the bundled skill from `internal/skills/bundled/agent-telegram`.
- Generate `llms.txt` through the repository's `agent-telegram llms-txt` command so its command catalog, schemas, safety classifications, and agent contract match the CLI.
- Expose the upstream repository URL beside the hosted artifacts so readers can inspect the canonical implementation.

## Public Artifact Layout

- `/llms.txt` — full generated CLI documentation for LLM/tool consumers.
- `/skills/agent-telegram/SKILL.md` — skill instructions.
- `/skills/agent-telegram/agents/openai.yaml` — skill UI metadata.

These files live under Vite's `public` directory and are copied verbatim to the deployed site. They do not depend on a network request during build or runtime.

## Landing Page Integration

Add an `#agents` section before the final install CTA. It explains two complementary discovery surfaces:

1. The bundled Codex skill teaches an agent the preferred workflow, authentication, safety gates, bot-flow commands, and observability practices. The primary command is `agent-telegram skills install agent-telegram`.
2. `/llms.txt` provides the complete generated command reference. The installed CLI can regenerate the same reference with `agent-telegram llms-txt`.

Add an “Agent docs” navigation item and direct, accessible links to the skill, `llms.txt`, and GitHub source. Keep the section static and lightweight; no new state or animation model is required beyond the existing section reveal behavior.

## Content Contract

The page and assets must communicate the upstream contract accurately:

- authentication is QR-only and local-browser based;
- agents should use `server ensure`, `--agent`, and a shared run ID for related commands;
- destructive and paid actions require user confirmation;
- `manifest`, `--schema`, receipts, traces, audit, and logs support discovery and debugging;
- redaction remains enabled by default.

## Verification

Verification tier: Tier 2 — route composition, visible content, navigation, and public artifact delivery change; renderer, canvas, runtime, and performance-sensitive paths do not.

Run:

- skill validation against the published skill directory;
- a source parity check for the copied skill files;
- targeted content tests for artifact URLs and canonical commands;
- `npm run verify:quick`;
- landing browser acceptance covering the new section and public files;
- `npm run build` and verify the files exist in `dist`.

Skip the full performance suite because this is a post-first-working static documentation/content addition with no renderer or interaction workload changes.
