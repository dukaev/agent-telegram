import * as React from "react";

import {
  createToolcraftPngExportCanvas,
  shouldIncludeToolcraftPreviewBackground,
  type ToolcraftMediaAsset,
  type ToolcraftState,
} from "@/toolcraft/runtime";
import { useToolcraft } from "@/toolcraft/runtime/react";

type DuckCardVariant = "cards" | "chat" | "hero" | "thinking";

type DuckCardSettings = {
  accent: string;
  background: string;
  caption: string;
  handle: string;
  headline: string;
  imageScale: number;
  variant: DuckCardVariant;
};

function readString(state: ToolcraftState, target: string, fallback: string): string {
  const value = state.values[target];
  return typeof value === "string" && value.trim() ? value : fallback;
}

function readSettings(state: ToolcraftState): DuckCardSettings {
  const variant = readString(state, "card.variant", "hero");

  return {
    accent: readString(state, "appearance.accent", "#229ED9"),
    background: readString(state, "appearance.background", "#F4F1E8"),
    caption: readString(
      state,
      "copy.caption",
      "Ваш дружелюбный AI-агент в Telegram",
    ),
    handle: readString(state, "copy.handle", "@duck_agent"),
    headline: readString(state, "copy.headline", "Утёнок думает за вас"),
    imageScale:
      typeof state.values["source.scale"] === "number"
        ? Number(state.values["source.scale"])
        : 104,
    variant:
      variant === "thinking" || variant === "chat" || variant === "cards"
        ? variant
        : "hero",
  };
}

function getSourceAsset(state: ToolcraftState): ToolcraftMediaAsset | undefined {
  return state.mediaAssets.find((asset) => asset.sourceTarget === "source.image");
}

function getImageTransform(asset: ToolcraftMediaAsset, scalePercent: number): string {
  const transform = asset.transform;
  const rotate = transform?.rotationDeg ?? 0;
  const scaleX = transform?.flipHorizontal ? -1 : 1;
  const scaleY = transform?.flipVertical ? -1 : 1;
  const scale = scalePercent / 100;
  return `scale(${scale * scaleX}, ${scale * scaleY}) rotate(${rotate}deg)`;
}

function TelegramMark(): React.JSX.Element {
  return (
    <span aria-hidden="true" className="duck-card__telegram-mark">
      ↗
    </span>
  );
}

export function DuckCardRenderer(): React.JSX.Element | null {
  const { state } = useToolcraft();
  const asset = getSourceAsset(state);

  if (!asset) {
    return null;
  }

  const settings = readSettings(state);
  const includeBackground = shouldIncludeToolcraftPreviewBackground({ state });
  const style = {
    "--duck-accent": settings.accent,
    "--duck-background": includeBackground ? settings.background : "transparent",
  } as React.CSSProperties;

  return (
    <article
      className={`duck-card duck-card--${settings.variant}`}
      data-duck-variant={settings.variant}
      data-toolcraft-product-output
      style={style}
    >
      <div className="duck-card__background" data-renderer-layer="backgroundLayer" />
      <div className="duck-card__orb duck-card__orb--one" />
      <div className="duck-card__orb duck-card__orb--two" />

      <div className="duck-card__content" data-renderer-layer="productForegroundLayer">
        <div className="duck-card__copy">
          <div className="duck-card__eyebrow">
            <TelegramMark />
            Telegram AI agent
          </div>
          <h1 className="duck-card__headline" data-toolcraft-product-text>
            {settings.headline}
          </h1>
          <p className="duck-card__caption" data-toolcraft-product-text>
            {settings.caption}
          </p>
          <div className="duck-card__cta-row">
            <span className="duck-card__cta">Открыть в Telegram</span>
            <span className="duck-card__handle" data-toolcraft-product-text>
              {settings.handle}
            </span>
          </div>
        </div>

        <div className="duck-card__visual">
          <div className="duck-card__image-shell">
            <img
              alt="Задумчивый жёлтый утёнок"
              className="duck-card__image"
              data-media-id={asset.id}
              data-media-source={asset.fileName}
              draggable={false}
              src={asset.dataUrl}
              style={{ transform: getImageTransform(asset, settings.imageScale) }}
            />
          </div>
          <div className="duck-card__status">thinking…</div>
        </div>

        {settings.variant === "thinking" ? (
          <div className="duck-card__thinking-line" aria-hidden="true">
            <span />
            <span />
            <span />
            <span />
            <span />
          </div>
        ) : null}

        {settings.variant === "chat" ? (
          <div className="duck-card__chat-stack">
            <div className="duck-card__bubble duck-card__bubble--question">
              Разберёшь задачу?
            </div>
            <div className="duck-card__bubble duck-card__bubble--answer">
              Уже думаю. Вернусь с готовым результатом ✦
            </div>
          </div>
        ) : null}

        {settings.variant === "cards" ? (
          <div className="duck-card__mini-grid">
            <div className="duck-card__mini-card">
              <b>01</b>
              Понимает контекст
            </div>
            <div className="duck-card__mini-card">
              <b>02</b>
              Делает работу
            </div>
            <div className="duck-card__mini-card">
              <b>03</b>
              Отвечает в Telegram
            </div>
          </div>
        ) : null}
      </div>
    </article>
  );
}

function roundedRect(
  context: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
): void {
  const r = Math.min(radius, width / 2, height / 2);
  context.beginPath();
  context.roundRect(x, y, width, height, r);
}

function loadImage(source: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.decoding = "async";
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error("Unable to decode the source image."));
    image.src = source;
  });
}

function drawCoverImage(
  context: CanvasRenderingContext2D,
  image: HTMLImageElement,
  asset: ToolcraftMediaAsset,
  box: { height: number; width: number; x: number; y: number },
  imageScale: number,
  radius: number,
): void {
  const rotation = asset.transform?.rotationDeg ?? 0;
  const quarterTurn = rotation === 90 || rotation === 270;
  const sourceWidth = quarterTurn ? image.naturalHeight : image.naturalWidth;
  const sourceHeight = quarterTurn ? image.naturalWidth : image.naturalHeight;
  const coverScale = Math.max(box.width / sourceWidth, box.height / sourceHeight);
  const scale = coverScale * (imageScale / 100);

  context.save();
  roundedRect(context, box.x, box.y, box.width, box.height, radius);
  context.clip();
  context.translate(box.x + box.width / 2, box.y + box.height / 2);
  context.rotate((rotation * Math.PI) / 180);
  context.scale(
    asset.transform?.flipHorizontal ? -1 : 1,
    asset.transform?.flipVertical ? -1 : 1,
  );
  context.drawImage(
    image,
    (-image.naturalWidth * scale) / 2,
    (-image.naturalHeight * scale) / 2,
    image.naturalWidth * scale,
    image.naturalHeight * scale,
  );
  context.restore();
}

function wrapText(
  context: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
): string[] {
  const words = text.split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let line = "";

  for (const word of words) {
    const candidate = line ? `${line} ${word}` : word;
    if (line && context.measureText(candidate).width > maxWidth) {
      lines.push(line);
      line = word;
    } else {
      line = candidate;
    }
  }

  if (line) lines.push(line);
  return lines;
}

function drawTextLines(
  context: CanvasRenderingContext2D,
  lines: readonly string[],
  x: number,
  y: number,
  lineHeight: number,
  maxLines = lines.length,
): void {
  lines.slice(0, maxLines).forEach((line, index) => {
    context.fillText(line, x, y + index * lineHeight);
  });
}

function drawDuckCard(
  context: CanvasRenderingContext2D,
  width: number,
  height: number,
  image: HTMLImageElement,
  asset: ToolcraftMediaAsset,
  settings: DuckCardSettings,
): void {
  const sx = width / 1920;
  const sy = height / 1080;
  const s = Math.min(sx, sy);
  const px = (value: number) => value * s;
  const left = Math.max(px(110), width * 0.058);
  const top = Math.max(px(82), height * 0.075);

  context.textBaseline = "alphabetic";
  context.fillStyle = "#161616";

  if (settings.variant === "thinking") {
    const imageSize = Math.min(width * 0.39, height * 0.63);
    const imageX = (width - imageSize) / 2;
    drawCoverImage(
      context,
      image,
      asset,
      { height: imageSize, width: imageSize, x: imageX, y: top + px(38) },
      settings.imageScale,
      px(72),
    );

    context.fillStyle = settings.accent;
    context.font = `700 ${px(24)}px Inter, sans-serif`;
    context.fillText("TELEGRAM AI AGENT", left, top);

    context.fillStyle = "#161616";
    context.font = `850 ${px(78)}px Inter, sans-serif`;
    context.textAlign = "center";
    const headlineLines = wrapText(context, settings.headline, width * 0.75);
    drawTextLines(context, headlineLines, width / 2, height - px(178), px(82), 2);
    context.font = `500 ${px(29)}px Inter, sans-serif`;
    context.fillStyle = "#505050";
    context.fillText(settings.handle, width / 2, height - px(72));
    context.textAlign = "left";
    return;
  }

  if (settings.variant === "chat") {
    context.fillStyle = "#111827";
    roundedRect(context, left, top, width - left * 2, height - top * 2, px(52));
    context.fill();

    const imageSize = Math.min(width * 0.31, height * 0.62);
    drawCoverImage(
      context,
      image,
      asset,
      {
        height: imageSize,
        width: imageSize,
        x: width - left - imageSize - px(48),
        y: (height - imageSize) / 2,
      },
      settings.imageScale,
      px(48),
    );

    context.fillStyle = "#ffffff";
    context.font = `820 ${px(72)}px Inter, sans-serif`;
    drawTextLines(context, wrapText(context, settings.headline, width * 0.44), left + px(60), top + px(150), px(78), 3);

    const bubbleX = left + px(60);
    const bubbleW = width * 0.38;
    context.fillStyle = "#ffffff";
    roundedRect(context, bubbleX, height * 0.49, bubbleW, px(86), px(26));
    context.fill();
    context.fillStyle = "#202634";
    context.font = `600 ${px(25)}px Inter, sans-serif`;
    context.fillText("Разберёшь задачу?", bubbleX + px(28), height * 0.49 + px(53));

    context.fillStyle = settings.accent;
    roundedRect(context, bubbleX + px(74), height * 0.60, bubbleW, px(112), px(28));
    context.fill();
    context.fillStyle = "#ffffff";
    context.font = `600 ${px(24)}px Inter, sans-serif`;
    drawTextLines(
      context,
      wrapText(context, "Уже думаю. Вернусь с готовым результатом ✦", bubbleW - px(54)),
      bubbleX + px(102),
      height * 0.60 + px(45),
      px(30),
      2,
    );
    context.fillStyle = "#99A3B7";
    context.font = `500 ${px(24)}px Inter, sans-serif`;
    context.fillText(settings.handle, bubbleX, height - top - px(32));
    return;
  }

  if (settings.variant === "cards") {
    const imageSize = Math.min(width * 0.30, height * 0.58);
    drawCoverImage(
      context,
      image,
      asset,
      { height: imageSize, width: imageSize, x: left, y: top },
      settings.imageScale,
      px(54),
    );
    context.fillStyle = settings.accent;
    roundedRect(context, left, top + imageSize + px(24), imageSize, px(110), px(32));
    context.fill();
    context.fillStyle = "#ffffff";
    context.font = `700 ${px(25)}px Inter, sans-serif`;
    context.fillText(settings.handle, left + px(32), top + imageSize + px(68));

    const contentX = left + imageSize + px(74);
    context.fillStyle = "#161616";
    context.font = `850 ${px(78)}px Inter, sans-serif`;
    drawTextLines(context, wrapText(context, settings.headline, width - contentX - left), contentX, top + px(70), px(84), 3);

    const cards = ["Понимает контекст", "Делает работу", "Отвечает в Telegram"];
    const cardW = (width - contentX - left - px(28)) / 2;
    cards.forEach((label, index) => {
      const x = contentX + (index % 2) * (cardW + px(28));
      const y = top + px(290) + Math.floor(index / 2) * px(190);
      context.fillStyle = index === 2 ? settings.accent : "#ffffff";
      roundedRect(context, x, y, index === 2 ? cardW * 1.35 : cardW, px(160), px(34));
      context.fill();
      context.fillStyle = index === 2 ? "#ffffff" : settings.accent;
      context.font = `800 ${px(23)}px Inter, sans-serif`;
      context.fillText(`0${index + 1}`, x + px(28), y + px(43));
      context.fillStyle = index === 2 ? "#ffffff" : "#242424";
      context.font = `650 ${px(25)}px Inter, sans-serif`;
      context.fillText(label, x + px(28), y + px(104));
    });
    return;
  }

  const photoW = Math.min(width * 0.39, height * 0.72);
  const photoX = width - left - photoW;
  const photoY = (height - photoW) / 2;
  context.fillStyle = `${settings.accent}22`;
  context.beginPath();
  context.arc(photoX + photoW / 2, photoY + photoW / 2, photoW * 0.62, 0, Math.PI * 2);
  context.fill();
  drawCoverImage(
    context,
    image,
    asset,
    { height: photoW, width: photoW, x: photoX, y: photoY },
    settings.imageScale,
    px(64),
  );

  context.fillStyle = settings.accent;
  context.font = `750 ${px(25)}px Inter, sans-serif`;
  context.fillText("↗  TELEGRAM AI AGENT", left, top + px(42));
  context.fillStyle = "#161616";
  context.font = `850 ${px(92)}px Inter, sans-serif`;
  drawTextLines(context, wrapText(context, settings.headline, width * 0.47), left, top + px(185), px(98), 3);
  context.fillStyle = "#555555";
  context.font = `500 ${px(31)}px Inter, sans-serif`;
  drawTextLines(context, wrapText(context, settings.caption, width * 0.43), left, height * 0.64, px(42), 3);
  context.fillStyle = settings.accent;
  roundedRect(context, left, height - top - px(92), px(330), px(78), px(39));
  context.fill();
  context.fillStyle = "#ffffff";
  context.font = `700 ${px(25)}px Inter, sans-serif`;
  context.fillText("Открыть в Telegram", left + px(34), height - top - px(43));
  context.fillStyle = "#3a3a3a";
  context.font = `600 ${px(24)}px Inter, sans-serif`;
  context.fillText(settings.handle, left + px(365), height - top - px(43));
}

function canvasToBlob(canvas: HTMLCanvasElement, mimeType: string): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error("Image encoding failed."))),
      mimeType,
      0.94,
    );
  });
}

export async function exportDuckCardImage(
  state: ToolcraftState,
  reportProgress: (progress: number) => void,
): Promise<void> {
  const asset = getSourceAsset(state);
  if (!asset) throw new Error("Add a source image before exporting.");

  reportProgress(0.12);
  const image = await loadImage(asset.dataUrl);
  const settings = readSettings(state);
  const includeBackground = Boolean(state.values["export.includeBackground"] ?? true);
  const resolution = readString(state, "export.image.resolution", "4k");

  reportProgress(0.38);
  const canvas = createToolcraftPngExportCanvas({
    background: settings.background,
    includeBackground: includeBackground,
    resolution: resolution,
    state,
    render: ({ context, cssHeight, cssWidth }) => {
      drawDuckCard(context, cssWidth, cssHeight, image, asset, settings);
    },
  });

  reportProgress(0.78);
  const format = readString(state, "export.image.format", "png");
  const mimeType = format === "jpg" ? "image/jpeg" : "image/png";
  const blob = await canvasToBlob(canvas, mimeType);
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `duck-agent-${settings.variant}.${format === "jpg" ? "jpg" : "png"}`;
  anchor.click();
  setTimeout(() => URL.revokeObjectURL(url), 0);
  reportProgress(1);
}
