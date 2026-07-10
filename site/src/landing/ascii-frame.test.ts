import { describe, expect, it } from "vitest";

import { ASCII_GLYPHS, computeAsciiGrid, mapAsciiPixel } from "./ascii-frame";

describe("ASCII frame mapping", () => {
  it("removes the near-white source background", () => {
    expect(mapAsciiPixel(250, 249, 247)).toBeNull();
  });

  it("removes neutral shadow wash without erasing dark outlines", () => {
    expect(mapAsciiPixel(155, 153, 151)).toBeNull();

    const outline = mapAsciiPixel(35, 22, 18);
    expect(outline).not.toBeNull();
    expect(outline?.color[0]).toBeGreaterThan(120);
  });

  it("maps saturated duck yellow to a dense visible glyph", () => {
    const yellow = mapAsciiPixel(248, 199, 47);

    expect(yellow).not.toBeNull();
    expect(ASCII_GLYPHS.indexOf(yellow!.glyph)).toBeGreaterThan(ASCII_GLYPHS.length * 0.7);
    expect(yellow?.alpha).toBeGreaterThan(0.8);
  });

  it("preserves monochrome halftone detail with a warm tonal grade", () => {
    const halftone = mapAsciiPixel(82, 82, 82);

    expect(halftone).not.toBeNull();
    expect(halftone!.color[0]).toBeGreaterThan(halftone!.color[1]);
    expect(halftone!.color[1]).toBeGreaterThan(halftone!.color[2]);
  });

  it("keeps the responsive grid inside its performance bounds", () => {
    expect(computeAsciiGrid(240, 688 / 640)).toEqual({ columns: 72, rows: 40 });
    expect(computeAsciiGrid(1200, 688 / 640)).toEqual({ columns: 144, rows: 80 });
  });
});
