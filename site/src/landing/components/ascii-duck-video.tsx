import { useEffect, useRef } from "react";

import { computeAsciiGrid, mapAsciiPixel } from "../ascii-frame";
import { useReducedMotionPreference } from "../use-reduced-motion";

const TARGET_FRAME_INTERVAL = 1000 / 12;
const SOURCE_ASPECT = 688 / 640;

export function AsciiDuckVideo(): React.JSX.Element {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const reduceMotion = useReducedMotionPreference();

  useEffect(() => {
    const canvas = canvasRef.current;
    const video = videoRef.current;
    if (!canvas || !video) return;

    const output = canvas.getContext("2d", { alpha: true });
    const sampleCanvas = document.createElement("canvas");
    const sample = sampleCanvas.getContext("2d", { alpha: false, willReadFrequently: true });
    if (!output || !sample) return;

    let animationFrame = 0;
    let lastDrawAt = -TARGET_FRAME_INTERVAL;
    let isVisible = true;
    let disposed = false;

    const drawFrame = (now = performance.now()) => {
      if (disposed || video.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) return;

      const bounds = canvas.getBoundingClientRect();
      if (bounds.width < 1 || bounds.height < 1) return;

      const pixelRatio = Math.min(window.devicePixelRatio || 1, 2);
      const outputWidth = Math.round(bounds.width * pixelRatio);
      const outputHeight = Math.round(bounds.height * pixelRatio);

      if (canvas.width !== outputWidth || canvas.height !== outputHeight) {
        canvas.width = outputWidth;
        canvas.height = outputHeight;
      } else {
        output.clearRect(0, 0, outputWidth, outputHeight);
      }

      const { columns, rows } = computeAsciiGrid(bounds.width, SOURCE_ASPECT);
      if (sampleCanvas.width !== columns || sampleCanvas.height !== rows) {
        sampleCanvas.width = columns;
        sampleCanvas.height = rows;
      }

      sample.drawImage(video, 0, 0, columns, rows);
      const pixels = sample.getImageData(0, 0, columns, rows).data;
      const cellWidth = outputWidth / columns;
      const cellHeight = outputHeight / rows;

      output.textAlign = "center";
      output.textBaseline = "middle";
      output.font = `600 ${Math.max(7, cellHeight * 0.92)}px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace`;

      for (let row = 0; row < rows; row += 1) {
        for (let column = 0; column < columns; column += 1) {
          const offset = (row * columns + column) * 4;
          const mapped = mapAsciiPixel(pixels[offset], pixels[offset + 1], pixels[offset + 2]);
          if (!mapped) continue;

          const [red, green, blue] = mapped.color;
          output.fillStyle = `rgba(${red}, ${green}, ${blue}, ${mapped.alpha})`;
          output.fillText(
            mapped.glyph,
            (column + 0.5) * cellWidth,
            (row + 0.5) * cellHeight,
          );
        }
      }

      canvas.dataset.columns = String(columns);
      canvas.dataset.rows = String(rows);
      canvas.dataset.frameTime = video.currentTime.toFixed(3);
    };

    const tick = (now: number) => {
      if (disposed) return;

      if (isVisible && now - lastDrawAt >= TARGET_FRAME_INTERVAL) {
        drawFrame(now);
        lastDrawAt = now;
      }

      animationFrame = window.requestAnimationFrame(tick);
    };

    const updatePlayback = () => {
      if (reduceMotion || !isVisible) {
        video.pause();
        if (reduceMotion && video.currentTime > 0.001) video.currentTime = 0;
        drawFrame();
        return;
      }

      void video.play().catch(() => drawFrame());
    };

    const observer = new IntersectionObserver(
      ([entry]) => {
        isVisible = entry.isIntersecting;
        updatePlayback();
      },
      { rootMargin: "120px 0px", threshold: 0.01 },
    );

    const resizeObserver = new ResizeObserver(() => drawFrame());
    const handleLoadedFrame = () => {
      drawFrame();
      updatePlayback();
    };

    observer.observe(canvas);
    resizeObserver.observe(canvas);
    video.addEventListener("loadeddata", handleLoadedFrame);
    video.addEventListener("seeked", handleLoadedFrame);
    animationFrame = window.requestAnimationFrame(tick);

    return () => {
      disposed = true;
      window.cancelAnimationFrame(animationFrame);
      observer.disconnect();
      resizeObserver.disconnect();
      video.removeEventListener("loadeddata", handleLoadedFrame);
      video.removeEventListener("seeked", handleLoadedFrame);
      video.pause();
    };
  }, [reduceMotion]);

  return (
    <div className="ascii-duck" data-ascii-duck>
      <video
        aria-hidden="true"
        className="ascii-duck__source"
        loop
        muted
        playsInline
        preload="auto"
        ref={videoRef}
        src="/duck-laptop.mp4?v=2cd449d9"
        tabIndex={-1}
      />
      <canvas
        aria-label="Animated ASCII illustration of a duck working on a laptop"
        className="ascii-duck__canvas"
        data-ascii-canvas
        ref={canvasRef}
        role="img"
      />
    </div>
  );
}
