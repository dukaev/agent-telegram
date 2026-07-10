export const ASCII_GLYPHS = ".,:;irsXA253hMHGS#9B&@";

export type AsciiPixel = {
  alpha: number;
  color: readonly [number, number, number];
  glyph: string;
};

const clampByte = (value: number): number => Math.max(0, Math.min(255, Math.round(value)));

function getPixelMetrics(red: number, green: number, blue: number) {
  const maximum = Math.max(red, green, blue);
  const minimum = Math.min(red, green, blue);
  const saturation = maximum - minimum;
  const luminance = red * 0.2126 + green * 0.7152 + blue * 0.0722;

  return { luminance, saturation };
}

export function mapAsciiPixel(red: number, green: number, blue: number): AsciiPixel | null {
  const { luminance, saturation } = getPixelMetrics(red, green, blue);
  const isNearWhite = luminance > 222 && saturation < 32;
  const isNeutralShadow = luminance > 98 && saturation < 24;

  if (isNearWhite || isNeutralShadow) return null;

  const strength = Math.min(255, Math.max(saturation * 1.2, (255 - luminance) * 0.92));
  if (strength < 25) return null;

  const glyphIndex = Math.min(
    ASCII_GLYPHS.length - 1,
    Math.floor((strength / 255) * ASCII_GLYPHS.length),
  );

  let color: readonly [number, number, number];
  if (saturation < 28) {
    const tone = strength / 255;
    color = [
      clampByte(114 + tone * 108),
      clampByte(78 + tone * 70),
      clampByte(48 + tone * 42),
    ];
  } else {
    const lift = Math.max(1, 108 / luminance);
    color = [clampByte(red * lift), clampByte(green * lift), clampByte(blue * lift)];
  }

  return {
    alpha: Math.min(0.98, 0.62 + strength / 620),
    color,
    glyph: ASCII_GLYPHS[glyphIndex],
  };
}

export function computeAsciiGrid(cssWidth: number, sourceAspect: number) {
  const columns = Math.max(72, Math.min(144, Math.round(cssWidth / 3.8)));
  const rows = Math.round(columns * sourceAspect * 0.52);

  return { columns, rows };
}
