# Ping-Pong Duck Video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hero duck asset with a seamless four-second loop that plays the first two seconds forward and then retraces them for two seconds.

**Architecture:** Bake the ping-pong sequence into the existing MP4 so the current native video loop, Canvas ASCII renderer, visibility handling, and reduced-motion behavior remain untouched. Generate and inspect a temporary output before replacing the committed binary.

**Tech Stack:** FFmpeg/ffprobe, React + Vite (unchanged), Playwright, Cloudflare Workers Builds.

---

## Task 1: Record the media decision and reference evidence

**Files:**
- Reference: `public/duck-laptop.mp4`
- Modify: `docs/toolcraft/agent-worklog.md`

- [ ] Classify the change as Tier 3 because the visible animation output changes while the renderer workload remains unchanged.
- [ ] Record the source stream metadata and storyboard samples at 0.0, 0.5, 1.0, 1.5, and approximately 2.0 seconds.
- [ ] Record the chosen baked-media approach and reject runtime negative playback rate and repeated backward seeks as less reliable alternatives.

## Task 2: Build the four-second ping-pong asset

**Files:**
- Reference: `public/duck-laptop.mp4`
- Create temporarily: `/tmp/duck-laptop-pingpong.mp4`
- Modify: `public/duck-laptop.mp4`

- [ ] Trim source time 0–2 seconds, split it, reverse the second copy, and concatenate the two halves.
- [ ] Encode H.264 at 24 fps, `yuv420p`, CRF 18, with fast-start metadata and no audio.
- [ ] Confirm the temporary file is 640×688, four seconds, 96 frames, and contains no audio stream.
- [ ] Generate a two-row contact sheet and compare mirrored frame pairs before replacing the committed asset.
- [ ] Replace only `public/duck-laptop.mp4`; do not modify React, Canvas, CSS, autoplay, visibility, reduced-motion, or accessibility code.

## Task 3: Verify the landing locally

**Files:**
- Test: `src/landing/content.test.ts`
- Test: `e2e/landing.spec.ts`

- [ ] Run `npm run verify:quick`.
- [ ] Run `npm run build`.
- [ ] Run `npx playwright test e2e/landing.spec.ts`.
- [ ] Inspect the local hero in a controlled browser through a full forward/reverse cycle.
- [ ] Confirm no console errors and no horizontal overflow.
- [ ] Record verification and explain why the full performance suite is skipped: the renderer, frame cap, sample density, and scheduling workload did not change.

## Task 4: Publish through the existing deployment path

**Files:**
- Modify: `docs/toolcraft/agent-worklog.md`

- [ ] Stage only the plan, design spec, video asset, and worklog; preserve unrelated working-tree changes.
- [ ] Commit the implementation on `main` and push it to `origin/main`.
- [ ] Wait for the native Cloudflare Workers Build for that commit to succeed.
- [ ] Verify production serves the new asset with a changed ETag and matching file checksum.
- [ ] Inspect the production hero and confirm the four-second ping-pong loop remains smooth.
