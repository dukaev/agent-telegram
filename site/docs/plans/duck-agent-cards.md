# Duck Agent Cards — Product Spec And Implementation Plan

## Product goal

Create one Toolcraft editor that presents all four requested concepts for `img.jpg` as switchable, exportable Telegram promo-card layouts. The default image remains attached source media and can be replaced by the user.

## Product behavior

- Canvas: editable 16:9 output, initially 1920×1080.
- Variants: `Hero`, `Thinking`, `Chat`, and `Cards` through one built-in segmented control.
- Source: one built-in image `fileDrop`, initialized from `/img.jpg`; upload replaces the source and runtime rotate/flip metadata reaches preview/export.
- Copy: editable headline, caption, and Telegram handle.
- Brand: editable accent color.
- Background: runtime `Include` switch plus editable product background color.
- Delivery: sticky `Export image` action; PNG/JPG format and 2K/4K/8K resolution controls use the standard image-export path.
- Persistence: intentionally omitted. Settings transfer remains runtime-owned; reload starts from defaults.
- Timeline: omitted because the product is still output with no transport or video export.
- Layers: omitted because there is one composed card output and one replaceable source image.
- Custom controls: none; every setting maps to a built-in control.

## Control Section Inventory

| Section | Entity / stage | Targets | Grouping reason |
| --- | --- | --- | --- |
| Layout | Card composition | `card.variant` | Selects one of the four requested layouts. |
| Photo | Source media attachment | `source.image` | Runtime media section owns upload, removal, reset, and transforms. |
| Source | Hero image crop | `source.scale` | Controls the crop scale of the attached source image. |
| Message | Promo copy | `copy.headline`, `copy.caption`, `copy.handle` | Edits the three visible text roles in every composition. |
| Brand | Accent role | `appearance.accent` | Applies the Telegram/brand accent across composition details. |
| Background | Product background | `export.includeBackground`, `appearance.background` | Required live-preview and export background behavior. |
| Image Export | Delivery settings | `export.image.format`, `export.image.resolution` | Controls exported image encoding and real output dimensions. |

## Renderer Technique Decision Matrix

- sourceRepresentation: Runtime media asset with data URL and transform metadata.
- productRepresentation: Low-count structured promo card with semantic text and one image.
- previewRenderer: DOM/CSS product renderer for crisp native text and responsive layout.
- exportRenderer: Canvas 2D composite at the requested Toolcraft export dimensions.
- rendererWorkload: `text-output` plus one decoded bitmap; no per-pixel processing.
- rendererStrategy: `mixed`.
- whyNotAlternativeStrategies: SVG would complicate bitmap crop/transform and wrapping; WebGL/WebGPU adds no value for one image and low-count geometry; Canvas-only preview would rasterize native text unnecessarily.
- fidelityRisks: Preview/export text wrapping and image crop must remain visually close; export uses deterministic font sizes and line breaking.
- performanceRisks: Image decode on source replacement and high-resolution export allocation; preview uses the browser-decoded asset and export decodes once per action.

## Renderer Layer Inventory

- `backgroundLayer`: configurable solid product background; preview visibility follows `shouldIncludeToolcraftPreviewBackground`.
- `productForegroundLayer`: card chrome, source image, headline, caption, handle, badges, and chat bubbles.
- `editingHandlesLayer`: none.
- `exportComposite`: Canvas 2D drawing of background (when included), image, card geometry, and text.

## Render Pipeline Inventory

1. `source-decode` — image decode; cache key is media asset id/data URL/transform; invalidated by `media-import` and media transform.
2. `text-layout` — DOM preview / Canvas export text layout; cache key is variant plus copy values and canvas dimensions; invalidated by copy, variant, or canvas-size changes.
3. `composition` — DOM/CSS preview composition; invalidated by variant, copy, accent, background, include-background, media, and viewport-size changes.
4. `export-composite` — Canvas 2D export-quality pass; invalidated on export by all product targets, media, canvas size, selected format, and selected resolution.

Interaction invalidation: control changes update only composition/text layout; viewport drag and zoom do not re-decode the image; media import/rotate/flip invalidates source decode; export builds a fresh export composite.

## Verification note

Verification tier: Tier 4

Reason: This turns a fresh neutral starter into the first working product version with schema controls, default media, a custom renderer, export, acceptance coverage, and performance declarations.

Run: `npm run ai:check`, `npm run verify:final`, first-working browser performance checkpoint (agent-browser preferred, Playwright fallback), and `npm run dev`.

Skip: Timeline, video-export, layer-interaction, canvas-handle, and animation checks because the product intentionally has none of those behaviors.

## Implementation plan

1. Copy `img.jpg` to `public/img.jpg` as the schema-backed default media asset.
2. Update `src/app/app-schema.ts` with editable output sizing, product sections, default media, built-in controls, toolbar, and sticky export action.
3. Add `src/app/duck-card-renderer.tsx` and export helpers for all four variants; compose them through the thin route using `ToolcraftApp`.
4. Add product styling in `src/styles.css` without changing copied Toolcraft runtime files.
5. Convert `src/app/app-acceptance.ts`, schema tests, browser tests, and performance config/tests from starter assertions to product observables.
6. Update `docs/toolcraft/agent-worklog.md` with the decision trail and exact verification evidence.
7. Run the Tier 4 checks, diagnose any failures through the required systematic-debugging workflow, verify the real UI in a browser, capture final evidence, and start the dev server.
