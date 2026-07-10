import { expect, test } from "@playwright/test";

import { AGENT_SETUP_PROMPT } from "../src/landing/content";

test("browser: landing tells the real-user Telegram story", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(/agent-telegram/);
  await expect(
    page.getByRole("heading", { name: /Your agent can use Telegram/i }),
  ).toBeVisible();
  await expect(page.getByText("TELEGRAM FOR AI AGENTS")).toBeVisible();
  await expect(page.locator("#top").getByText("01 / PASTE INTO YOUR AGENT")).toBeVisible();
  await expect(page.locator("#top").getByText(/Scan one QR code/)).toBeVisible();
  await expect(page.getByRole("link", { name: "View on GitHub" })).toHaveAttribute(
    "href",
    "https://github.com/dukaev/agent-telegram",
  );
  const asciiCanvas = page.locator("#top [data-ascii-canvas]");
  await expect(asciiCanvas).toBeVisible();
  await expect(page.locator("#top").getByText("LIVE USER SESSION")).toHaveCount(0);
  await expect
    .poll(async () => Number(await asciiCanvas.getAttribute("data-frame-time")))
    .toBeGreaterThan(0.1);
  await expect
    .poll(async () => Number(await asciiCanvas.getAttribute("data-columns")))
    .toBeGreaterThanOrEqual(110);
  await expect(page.locator(".hero__duck-stage")).toBeVisible();
  await expect(page.locator(".ascii-frame")).toHaveCount(0);
  await expect(page.getByText("DUCK.ASCII")).toHaveCount(0);
  await expect(page.getByText("12 FPS")).toHaveCount(0);
  await expect(page.getByText("CANVAS 2D · LOCAL MEDIA")).toHaveCount(0);

  const visibleAsciiPixels = await asciiCanvas.evaluate((canvas: HTMLCanvasElement) => {
    const context = canvas.getContext("2d");
    if (!context) return 0;
    const pixels = context.getImageData(0, 0, canvas.width, canvas.height).data;
    let visible = 0;
    for (let index = 3; index < pixels.length; index += 4) {
      if (pixels[index] > 0) visible += 1;
    }
    return visible;
  });
  expect(visibleAsciiPixels).toBeGreaterThan(1_000);

  await expect(
    page.getByRole("heading", { name: /Your endpoint can pass while the conversation is broken/i }),
  ).toBeVisible();
});

test("browser: outside-in section explains the evidence gap without controls", async ({ page }) => {
  await page.goto("/");

  const section = page.locator("#testing");
  await expect(section.getByText("YOUR BACKEND CAN PROVE")).toBeVisible();
  await expect(section.getByText("AGENT-TELEGRAM CAN VERIFY")).toBeVisible();
  await expect(section.getByText("The handler ran")).toBeVisible();
  await expect(section.getByText("The message appeared in the right chat")).toBeVisible();
  await expect(section.getByText("Structured receipts")).toBeVisible();
  await expect(section.getByRole("tab")).toHaveCount(0);
  await expect(section.getByRole("button")).toHaveCount(0);
});

test("browser: agent setup prompt copies with visible feedback", async ({ context, page }) => {
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.goto("/");

  const hero = page.locator("#top");
  const copyButton = hero.getByRole("button", { name: "Copy agent setup prompt" });
  await copyButton.click();

  await expect(hero.getByRole("button", { name: "Agent setup prompt copied" })).toBeVisible();
  await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe(AGENT_SETUP_PROMPT);
});

test("browser: mobile navigation works without horizontal overflow", async ({ page }) => {
  await page.setViewportSize({ height: 844, width: 390 });
  await page.goto("/");

  await page.getByRole("button", { name: "Open navigation" }).click();
  await expect(page.getByRole("navigation", { name: "Mobile navigation" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Bot testing" })).toBeVisible();
  await expect(page.getByRole("link", { name: "See the 2-step setup" })).toBeVisible();

  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(1);

  const asciiBounds = await page.locator("[data-ascii-canvas]").boundingBox();
  expect(asciiBounds).not.toBeNull();
  expect(asciiBounds!.x).toBeGreaterThanOrEqual(0);
  expect(asciiBounds!.x + asciiBounds!.width).toBeLessThanOrEqual(390);
});

test("browser: reduced motion keeps all content and disables demo autoplay", async ({ page }) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");

  const asciiCanvas = page.locator("[data-ascii-canvas]");
  await expect(asciiCanvas).toBeVisible();
  await expect.poll(() => asciiCanvas.getAttribute("data-frame-time")).toBe("0.000");
  await page.waitForTimeout(350);
  await expect(asciiCanvas).toHaveAttribute("data-frame-time", "0.000");
  await expect(page.locator("#testing").getByRole("tab")).toHaveCount(0);
  await expect(
    page.getByRole("heading", { name: /Your endpoint can pass while the conversation is broken/i }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { name: /Powerful enough to act/i })).toBeVisible();
  await expect(page.getByRole("heading", { name: /Install the CLI. Add the skill/i })).toBeVisible();
});
