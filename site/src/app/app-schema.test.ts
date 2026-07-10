import { describe, expect, it } from "vitest";

import { appPerformance } from "./app-performance";
import { appSchema } from "./app-schema";

function getControl(target: string) {
  return (appSchema.panels.controls?.sections ?? [])
    .flatMap((section) => Object.values(section.controls))
    .find((control) => control.target === target);
}

describe("Duck Agent Cards schema", () => {
  it("keeps the required Toolcraft runtime shell", () => {
    expect(appSchema.canvas).toMatchObject({
      draggable: true,
      enabled: true,
      size: { height: 1080, unit: "px", width: 1920 },
      sizing: { mode: "editable-output" },
      upload: true,
    });
    expect(appSchema.panels.controls?.sections[0]?.title).toBe("Setup");
    expect(appSchema.panels.layers).toBeUndefined();
    expect(appSchema.panels.timeline).toBeUndefined();
    expect(appSchema.toolbar).toEqual({
      history: true,
      radar: true,
      theme: true,
      zoom: true,
    });
  });

  it("publishes all four requested layout variants", () => {
    expect(getControl("card.variant")).toMatchObject({
      defaultValue: "hero",
      type: "segmented",
    });
    expect(getControl("card.variant")?.options?.map((option) => option.value)).toEqual([
      "hero",
      "thinking",
      "chat",
      "cards",
    ]);
  });

  it("attaches img.jpg as resettable source media", () => {
    expect(appSchema.media.defaultAssets).toEqual([
      expect.objectContaining({
        dataUrl: "/img.jpg",
        fileName: "img.jpg",
        sourceTarget: "source.image",
      }),
    ]);
    expect(getControl("source.image")).toMatchObject({ assetKind: "image", type: "fileDrop" });
    expect(getControl("source.scale")).toMatchObject({ defaultValue: 104, max: 140, min: 80 });
  });

  it("keeps background and image export controls in contract order", () => {
    const titles = appSchema.panels.controls?.sections.map((section) => section.title);
    expect(titles).toEqual([
      "Setup",
      "Layout",
      "Photo",
      "Source",
      "Message",
      "Brand",
      "Background",
      "Image Export",
      "Export",
    ]);
    expect(getControl("export.includeBackground")).toMatchObject({
      defaultValue: true,
      label: "Include",
      type: "switch",
    });
    expect(getControl("export.image.resolution")?.options?.map((option) => option.value)).toEqual([
      "2k",
      "4k",
      "8k",
    ]);
  });

  it("declares the custom DOM preview and Canvas export workload", () => {
    expect(appPerformance).toMatchObject({
      rendererStrategy: "dom",
      rendererWorkload: "text-output",
      usesCustomRenderer: true,
      workloadTargets: ["source.scale", "export.image.resolution"],
    });
  });
});
