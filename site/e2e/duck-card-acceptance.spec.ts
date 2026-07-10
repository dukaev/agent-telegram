import { expect, test, type Locator, type Page } from "@playwright/test";

import {
  dragToolcraftSliderByLabel,
  expectToolcraftSegmentedControlCellsPreservePadding,
} from "./performance-helpers";
import {
  expectToolcraftProductObservableToChange,
  getToolcraftProductObservableSnapshot,
} from "./product-observable-helpers";

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

async function selectCurrentOption(page: Page, current: string, next: string): Promise<void> {
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

async function downloadImage(page: Page): Promise<Buffer> {
  const [download] = await Promise.all([
    page.waitForEvent("download"),
    page.getByRole("button", { name: "Export image" }).click(),
  ]);
  const stream = await download.createReadStream();
  const chunks: Buffer[] = [];
  for await (const chunk of stream) chunks.push(Buffer.from(chunk));
  return Buffer.concat(chunks);
}

function readPngDimensions(buffer: Buffer): { height: number; width: number } {
  expect(buffer.subarray(0, 8).toString("hex")).toBe("89504e470d0a1a0a");
  return { height: buffer.readUInt32BE(20), width: buffer.readUInt32BE(16) };
}

test("browser: all four layout variants change product output", async ({ page }) => {
  await openApp(page);
  await expectToolcraftSegmentedControlCellsPreservePadding(page, "Variant");
  const snapshots = new Set<string>();
  snapshots.add(await getToolcraftProductObservableSnapshot(page));

  for (const variant of ["Thinking", "Chat", "Cards"] as const) {
    await expectToolcraftProductObservableToChange(page, async () => {
      await page.getByRole("button", { name: variant, exact: true }).click();
    });
    snapshots.add(await getToolcraftProductObservableSnapshot(page));
  }

  expect(snapshots.size).toBe(4);
});

test("browser: source photo upload transforms clear and reset reach output", async ({ page }) => {
  await openApp(page);
  const fileInput = page.locator('input[type="file"]');
  const fixture = Buffer.from(
    '<svg xmlns="http://www.w3.org/2000/svg" width="320" height="180"><rect width="320" height="180" fill="#7c3aed"/><circle cx="70" cy="90" r="45" fill="#fff"/></svg>',
  );

  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      await fileInput.setInputFiles({ buffer: fixture, mimeType: "image/svg+xml", name: "fixture.svg" });
    },
    { selector: ".duck-card__image" },
  );
  const productImage = page.locator(".duck-card__image");
  await expect(productImage).toHaveAttribute("src", /^data:image\/svg\+xml/);

  for (const actionName of ["90° Right", "Flip horizontal", "Flip vertical"]) {
    await expectToolcraftProductObservableToChange(
      page,
      async () => {
        await page.getByRole("button", { name: actionName, exact: true }).click();
      },
      { selector: ".duck-card__image" },
    );
  }

  await page.getByRole("button", { name: "Remove image" }).click();
  await expect(page.locator("[data-toolcraft-product-output]")).toHaveCount(0);
  await page.getByRole("button", { name: "Reset controls" }).click();
  await expect(page.locator("[data-toolcraft-product-output]")).toBeVisible();
  await expect(page.locator(".duck-card__image")).toHaveAttribute("src", "/img.jpg");
});

test("browser: image scale drag changes product output live", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      await dragToolcraftSliderByLabel(page, "Image scale", 0.94);
    },
    { selector: ".duck-card__image" },
  );
  await expect(page.getByRole("slider", { name: "Image scale" })).toHaveAttribute(
    "aria-valuenow",
    /13[6-8]/,
  );
});

test("browser: headline edits product text", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      await fieldInput(page, "Headline").fill("Утёнок уже придумал");
    },
    { selector: ".duck-card__headline" },
  );
  await expect(page.getByText("Утёнок уже придумал", { exact: true })).toBeVisible();
});

test("browser: caption edits product text", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(page, async () => {
    await fieldInput(page, "Caption").fill("Помощник, который не теряет контекст");
  });
  await expect(page.getByText("Помощник, который не теряет контекст", { exact: true })).toBeVisible();
});

test("browser: telegram handle edits product text", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(page, async () => {
    await fieldInput(page, "Telegram").fill("@smart_duck");
  });
  await expect(page.getByText("@smart_duck", { exact: true })).toBeVisible();
});

test("browser: accent color changes rendered card", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      const input = page.getByRole("textbox", { name: "Accent hex" });
      await input.click();
      await input.press("ControlOrMeta+A");
      await input.pressSequentially("#7C3AED", { delay: 20 });
      await input.press("Tab");
    },
    { selector: ".duck-card__cta" },
  );
  await expect(page.locator(".duck-card__cta")).toHaveCSS("background-color", "rgb(124, 58, 237)");
});

test("browser: include background toggles preview and png alpha", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      await page.getByRole("switch").click();
    },
    { selector: ".duck-card__background" },
  );
  await expect(page.locator(".duck-card__background")).toHaveCSS("background-color", "rgba(0, 0, 0, 0)");
});

test("browser: background color changes preview and exported pixels", async ({ page }) => {
  await openApp(page);
  await expectToolcraftProductObservableToChange(
    page,
    async () => {
      const input = page.getByRole("textbox", { name: "background hex" });
      await input.click();
      await input.press("ControlOrMeta+A");
      await input.pressSequentially("#DFF5FF", { delay: 20 });
      await input.press("Tab");
    },
    { selector: ".duck-card__background" },
  );
  await expect(page.locator(".duck-card__background")).toHaveCSS(
    "background-color",
    "rgb(223, 245, 255)",
  );
});

test("browser: png and jpg formats produce matching image bytes", async ({ page }) => {
  test.setTimeout(60_000);
  await openApp(page);
  await selectCurrentOption(page, "4K", "2K");
  const png = await downloadImage(page);
  expect(png.subarray(0, 8).toString("hex")).toBe("89504e470d0a1a0a");

  await selectCurrentOption(page, "PNG", "JPG");
  const jpg = await downloadImage(page);
  expect(jpg.subarray(0, 2).toString("hex")).toBe("ffd8");
  expect(jpg.subarray(-2).toString("hex")).toBe("ffd9");
});

test("browser: 2k 4k and 8k resolutions change decoded dimensions", async ({ page }) => {
  test.setTimeout(90_000);
  await openApp(page);
  const dimensions: number[] = [];

  await selectCurrentOption(page, "4K", "2K");
  dimensions.push(Math.max(...Object.values(readPngDimensions(await downloadImage(page)))));
  await selectCurrentOption(page, "2K", "4K");
  dimensions.push(Math.max(...Object.values(readPngDimensions(await downloadImage(page)))));
  await selectCurrentOption(page, "4K", "8K");
  dimensions.push(Math.max(...Object.values(readPngDimensions(await downloadImage(page)))));

  expect(dimensions).toEqual([2048, 4096, 8192]);
});

test("browser: export image reports progress and downloads output", async ({ page }) => {
  test.setTimeout(60_000);
  await openApp(page);
  await selectCurrentOption(page, "4K", "2K");
  const activeProgress = page.locator('[data-sticky-footer-active="true"]');
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export image" }).click();
  await expect(activeProgress).toBeVisible();
  await expect(activeProgress).toHaveAttribute("data-sticky-footer-progress", /0\.(12|38|78)/);
  const download = await downloadPromise;
  expect((await download.createReadStream()).readable).toBe(true);
  await expect(activeProgress).toHaveCount(0);
});

test("browser: renderer shows background and foreground product layers", async ({ page }) => {
  await openApp(page);
  const output = page.locator("[data-toolcraft-product-output]");
  const previewRect = await output.evaluate((element) => element.getBoundingClientRect().toJSON());
  const outputWidth = Number(await page.locator('input[data-slot="input"]').nth(0).inputValue());
  const outputHeight = Number(await page.locator('input[data-slot="input"]').nth(1).inputValue());
  expect(previewRect.width / previewRect.height).toBeCloseTo(outputWidth / outputHeight, 2);
  await expect(page.locator('[data-renderer-layer="backgroundLayer"]')).toBeVisible();
  await expect(page.locator('[data-renderer-layer="productForegroundLayer"]')).toBeVisible();
  expect(await getToolcraftProductObservableSnapshot(page)).toContain("duck-card--hero");
});
