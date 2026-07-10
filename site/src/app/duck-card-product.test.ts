import { describe, expect, it } from "vitest";

import { appAcceptance } from "./app-acceptance";
import { appPerformance } from "./app-performance";

const acceptanceById = new Map(appAcceptance.map((entry) => [entry.id, entry]));
const performanceById = new Map(appPerformance.scenarios.map((scenario) => [scenario.id, scenario]));

describe("Duck Agent Cards automated product evidence", () => {
  it("layout variants map to distinct product output", () => {
    expect(acceptanceById.get("card.variant")?.optionCoverage).toHaveLength(4);
  });
  it("source media lifecycle reaches preview and export", () => {
    expect(acceptanceById.get("source.image")?.evidence).toBe("media-lifecycle");
  });
  it("image scale changes live source crop", () => {
    expect(acceptanceById.get("source.scale")?.target).toBe("source.scale");
  });
  it("headline text reaches every layout", () => {
    expect(acceptanceById.get("copy.headline")).toBeDefined();
  });
  it("caption text reaches product output", () => {
    expect(acceptanceById.get("copy.caption")).toBeDefined();
  });
  it("telegram handle reaches product output", () => {
    expect(acceptanceById.get("copy.handle")).toBeDefined();
  });
  it("accent color reaches product composition", () => {
    expect(acceptanceById.get("appearance.accent")?.evidence).toBe("rendered-pixels");
  });
  it("include background controls preview and export alpha", () => {
    expect(acceptanceById.get("export.includeBackground")?.expectedObservable).toMatch(/alpha/i);
  });
  it("background color reaches preview and export", () => {
    expect(acceptanceById.get("appearance.background")?.evidence).toBe("rendered-pixels");
  });
  it("image format controls exported mime and extension", () => {
    expect(acceptanceById.get("export.image.format")?.optionCoverage).toEqual(["png", "jpg"]);
  });
  it("image resolution controls real exported dimensions", () => {
    expect(acceptanceById.get("export.image.resolution")?.optionCoverage).toEqual([
      "2k",
      "4k",
      "8k",
    ]);
  });
  it("export action returns real async image bytes", () => {
    expect(acceptanceById.get("actions.output")?.actionCoverage).toEqual(["export.png"]);
  });
  it("custom renderer exposes semantic product layers", () => {
    expect(acceptanceById.get("renderer.card")?.kind).toBe("runtime");
  });
});

describe("Duck Agent Cards automated performance evidence", () => {
  it("perf: initial preview renders within budget", () => {
    expect(performanceById.get("initial-preview")?.interaction).toBe("preview-render");
  });
  it("perf: variant-change updates within budget", () => {
    expect(performanceById.get("variant-change")?.target).toBe("card.variant");
  });
  it("perf: source image import stays responsive", () => {
    expect(performanceById.get("source-media-import")?.stressFixture?.kind).toBe("media");
  });
  it("perf: source scale drag stays responsive", () => {
    expect(performanceById.get("source-scale-drag")?.interaction).toBe("control-drag");
  });
  it("perf: headline-change updates within budget", () => {
    expect(performanceById.get("headline-change")?.target).toBe("copy.headline");
  });
  it("perf: caption-change updates within budget", () => {
    expect(performanceById.get("caption-change")?.target).toBe("copy.caption");
  });
  it("perf: handle-change updates within budget", () => {
    expect(performanceById.get("handle-change")?.target).toBe("copy.handle");
  });
  it("perf: accent-change updates within budget", () => {
    expect(performanceById.get("accent-change")?.target).toBe("appearance.accent");
  });
  it("perf: include-background-change updates within budget", () => {
    expect(performanceById.get("include-background-change")?.target).toBe(
      "export.includeBackground",
    );
  });
  it("perf: background-change updates within budget", () => {
    expect(performanceById.get("background-change")?.target).toBe("appearance.background");
  });
  it("perf: format-change updates within budget", () => {
    expect(performanceById.get("format-change")?.target).toBe("export.image.format");
  });
  it("perf: export resolution selection stays responsive", () => {
    expect(performanceById.get("resolution-change")?.stressFixture?.value).toBe("8k");
  });
  it("perf: 8K export completes within budget", () => {
    expect(performanceById.get("export-image")?.interaction).toBe("export-copy");
  });
  it("perf: zoom remains smooth with 4K source", () => {
    expect(performanceById.get("viewport-zoom")?.interaction).toBe("viewport-zoom-stress");
  });
  it("perf: viewport remains stable while editing", () => {
    expect(performanceById.get("viewport-stability")?.interaction).toBe("viewport-stability");
  });
});
