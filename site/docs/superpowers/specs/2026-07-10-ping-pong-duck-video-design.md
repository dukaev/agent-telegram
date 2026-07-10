# Ping-Pong Duck Video Design

## Goal

Make the hero duck play the first two seconds forward and then the same motion backward for two seconds, repeating as a seamless four-second loop.

## Approach

Bake the ping-pong motion into `public/duck-laptop.mp4`. The first half contains source time `0.000–2.000` seconds in its original order. The second half contains the same decoded frames in reverse order. The browser continues to play the resulting four-second asset forward with the existing native `loop` behavior.

This is preferred over runtime reverse playback. Negative HTML video playback rates are not consistently supported, and repeatedly seeking backward through the current H.264 file would be less efficient and more likely to stutter because the source contains only one keyframe at the beginning.

## Output Contract

- Path remains `public/duck-laptop.mp4`.
- Duration is four seconds: two seconds forward plus two seconds backward.
- Resolution remains `640 × 688`.
- Frame rate remains `24 fps`.
- Codec remains H.264 with `yuv420p` pixel format and web fast-start metadata.
- The output has no audio track.
- The first frame remains unchanged so reduced-motion output does not change.
- The turn at two seconds and the loop boundary may repeat one endpoint frame for a single 24 fps frame; no longer pause is introduced.

## Application Scope

No React, Canvas, CSS, autoplay, visibility, reduced-motion, accessibility, or deployment logic changes. `AsciiDuckVideo` continues to render `/duck-laptop.mp4?v=2cd449d9` through the existing `<video loop>` element. The production asset currently uses `cache-control: public, max-age=0, must-revalidate`, so the same URL will revalidate against the new deployment and does not require a query-string change.

## Verification

1. Inspect the output with `ffprobe` and confirm H.264, `640 × 688`, 24 fps, four-second duration, 96 frames, and no audio stream.
2. Render a contact sheet covering both halves and verify that the second half retraces the first half in reverse.
3. Compare paired frames around `0.5 ↔ 3.5` seconds and `1.5 ↔ 2.5` seconds to confirm visual correspondence.
4. Run `npm run verify:quick`, `npm run build`, and the landing Playwright suite without changing their existing product contracts.
5. Inspect the hero in a controlled browser, observe at least one two-second turning point, and confirm that the ASCII renderer remains smooth with no console errors or horizontal overflow.
6. Commit and push the asset to `main`, then confirm the native Cloudflare Workers Build succeeds and production serves the new asset ETag.

## Acceptance Criteria

- The duck moves forward for two seconds and retraces the same motion for two seconds.
- The four-second ping-pong repeats continuously without a visible hard cut.
- The hero layout and ASCII treatment are unchanged.
- Reduced-motion users still receive the unchanged first frame.
- Production updates through the existing automatic `main` deployment.
