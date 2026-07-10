import { expect, test, type Locator, type Page } from "@playwright/test";

import { appPerformance } from "../src/app/app-performance";
import {
  dragToolcraftSliderByLabel,
  dragToolcraftSliderToPerformanceStressValue,
  expectToolcraftCanvasViewportStable,
  expectToolcraftScenarioPerformanceBudget,
  getToolcraftPerformanceStressValue,
  getToolcraftPerformanceWorkloadValue,
  measureToolcraftInteraction,
  zoomToolcraftCanvasViewport,
} from "./performance-helpers";

async function openApp(page: Page): Promise<void> {
  await page.goto("/");
  await expect(page.locator("[data-toolcraft-product-output]")).toBeVisible();
}

function fieldInput(page: Page, label: string): Locator {
  return page
    .locator('[role="group"]')
    .filter({ has: page.getByText(label, { exact: true }) })
    .locator('input[data-slot="input"]')
    .first();
}

async function applyMediaFixture(
  page: Page,
  dimensions: { height: number; width: number },
): Promise<void> {
  const svg = Buffer.from(
    `<svg xmlns="http://www.w3.org/2000/svg" width="${dimensions.width}" height="${dimensions.height}"><rect width="100%" height="100%" fill="#7c3aed"/><circle cx="35%" cy="50%" r="18%" fill="#fff"/></svg>`,
  );
  await page.locator('input[type="file"]').setInputFiles({
    buffer: svg,
    mimeType: "image/svg+xml",
    name: `stress-${dimensions.width}x${dimensions.height}.svg`,
  });
  await expect(page.locator(".duck-card__image")).toHaveAttribute("src", /^data:image\/svg\+xml/);
}

async function chooseCurrentOption(page: Page, current: string, next: string): Promise<void> {
  const combobox = page.getByRole("combobox", { name: current, exact: true });
  await combobox.click();
  const direction =
    (current === "4K" && next === "2K") || (current === "JPG" && next === "PNG")
      ? "ArrowUp"
      : "ArrowDown";
  await combobox.press(direction);
  await combobox.press("Enter");
  await expect(page.getByRole("combobox", { name: next, exact: true })).toBeVisible();
}

test("browser perf: initial preview renders within budget", async ({ page }) => {
  const stress = getToolcraftPerformanceStressValue<{ height: number; width: number }>(
    appPerformance,
    "initial-preview",
  );
  const result = await measureToolcraftInteraction(page, async () => {
    await page.goto("/");
    await applyMediaFixture(page, stress);
    await expect(page.locator("[data-toolcraft-product-output]")).toBeVisible();
  });
  expect(result.sampleCount).toBeGreaterThan(0);
  expectToolcraftScenarioPerformanceBudget(
    { ...result, previewMs: result.durationMs },
    appPerformance,
    "initial-preview",
  );
});

test("browser perf: variant-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await page.getByRole("button", { name: "Chat", exact: true }).click();
  });
  await expect(page.locator('[data-duck-variant="chat"]')).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "variant-change");
});

test("browser perf: source image import stays responsive", async ({ page }) => {
  await openApp(page);
  const stress = getToolcraftPerformanceStressValue<{ height: number; width: number }>(
    appPerformance,
    "source-media-import",
  );
  const result = await measureToolcraftInteraction(page, async () => {
    await applyMediaFixture(page, stress);
  });
  await expect(page.locator(".duck-card__image")).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "source-media-import");
});

test("browser perf: source scale drag stays responsive", async ({ page }) => {
  await openApp(page);
  const workload = getToolcraftPerformanceWorkloadValue<{ height: number; width: number }>(
    appPerformance,
    "source-scale-drag",
  );
  await applyMediaFixture(page, workload);
  const stress = getToolcraftPerformanceStressValue<number>(appPerformance, "source-scale-drag");
  expect(stress).toBe(140);
  const result = await measureToolcraftInteraction(page, async () => {
    await dragToolcraftSliderByLabel(page, "Image scale", 0.86);
    await dragToolcraftSliderToPerformanceStressValue(
      page,
      "Image scale",
      appPerformance,
      "source-scale-drag",
    );
  });
  await expect(page.getByRole("slider", { name: "Image scale" })).toHaveAttribute(
    "aria-valuenow",
    "140",
  );
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "source-scale-drag");
});

test("browser perf: headline-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await fieldInput(page, "Headline").fill("Быстрый утёнок");
  });
  await expect(page.getByText("Быстрый утёнок", { exact: true })).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "headline-change");
});

test("browser perf: caption-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await fieldInput(page, "Caption").fill("Ответ без долгого ожидания");
  });
  await expect(page.getByText("Ответ без долгого ожидания", { exact: true })).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "caption-change");
});

test("browser perf: handle-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await fieldInput(page, "Telegram").fill("@fast_duck");
  });
  await expect(page.getByText("@fast_duck", { exact: true })).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "handle-change");
});

test("browser perf: accent-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    const input = page.getByRole("textbox", { name: "Accent hex" });
    await input.click();
    await input.press("ControlOrMeta+A");
    await input.pressSequentially("#7C3AED", { delay: 20 });
    await input.press("Tab");
  });
  await expect(page.locator(".duck-card__cta")).toHaveCSS(
    "background-color",
    "rgb(124, 58, 237)",
  );
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "accent-change");
});

test("browser perf: include-background-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await page.getByRole("switch").click();
  });
  await expect(page.getByRole("switch")).not.toBeChecked();
  expectToolcraftScenarioPerformanceBudget(
    result,
    appPerformance,
    "include-background-change",
  );
});

test("browser perf: background-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    const input = page.getByRole("textbox", { name: "background hex" });
    await input.click();
    await input.press("ControlOrMeta+A");
    await input.pressSequentially("#DFF5FF", { delay: 20 });
    await input.press("Tab");
  });
  await expect(page.locator(".duck-card__background")).toHaveCSS(
    "background-color",
    "rgb(223, 245, 255)",
  );
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "background-change");
});

test("browser perf: format-change updates within budget", async ({ page }) => {
  await openApp(page);
  const result = await measureToolcraftInteraction(page, async () => {
    await chooseCurrentOption(page, "PNG", "JPG");
  });
  await expect(page.getByRole("combobox", { name: "JPG", exact: true })).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "format-change");
});

test("browser perf: export resolution selection stays responsive", async ({ page }) => {
  await openApp(page);
  const workload = getToolcraftPerformanceWorkloadValue<{ height: number; width: number }>(
    appPerformance,
    "resolution-change",
  );
  await applyMediaFixture(page, workload);
  const stress = getToolcraftPerformanceStressValue<string>(appPerformance, "resolution-change");
  const result = await measureToolcraftInteraction(page, async () => {
    await chooseCurrentOption(page, "4K", stress.toUpperCase());
  });
  await expect(page.getByRole("combobox", { name: "8K", exact: true })).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "resolution-change");
});

test("browser perf: 8K export completes within budget", async ({ page }) => {
  test.setTimeout(60_000);
  await openApp(page);
  const stress = getToolcraftPerformanceStressValue<string>(appPerformance, "export-image");
  await chooseCurrentOption(page, "4K", stress.toUpperCase());
  const downloadPromise = page.waitForEvent("download");
  const result = await measureToolcraftInteraction(page, async () => {
    await page.getByRole("button", { name: "Export image" }).click();
    await downloadPromise;
  });
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toMatch(/\.png$/);
  expectToolcraftScenarioPerformanceBudget(
    { ...result, exportMs: result.durationMs },
    appPerformance,
    "export-image",
  );
});

test("browser perf: zoom remains smooth with 4K source", async ({ page }) => {
  await openApp(page);
  const stress = getToolcraftPerformanceStressValue<{ height: number; width: number }>(
    appPerformance,
    "viewport-zoom",
  );
  await applyMediaFixture(page, stress);
  const result = await measureToolcraftInteraction(page, async () => {
    await zoomToolcraftCanvasViewport(page, 2);
  });
  await expect(page.locator("[data-toolcraft-product-output]")).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "viewport-zoom");
});

test("browser perf: viewport remains stable while editing", async ({ page }) => {
  await openApp(page);
  const result = await expectToolcraftCanvasViewportStable(page, async () => {
    await fieldInput(page, "Headline").fill("Стабильный canvas");
  });
  await expect(page.locator("[data-toolcraft-product-output]")).toBeVisible();
  expectToolcraftScenarioPerformanceBudget(result, appPerformance, "viewport-stability");
});
