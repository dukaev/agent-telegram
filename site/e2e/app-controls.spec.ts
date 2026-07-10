import { expect, test } from "@playwright/test";

test("browser: Duck Agent Cards opens with product controls and default media", async ({ page }) => {
  await page.goto("/");

  await expect(page.locator('[data-slot="toolcraft-runtime-app"]')).toBeVisible();
  await expect(page.getByRole("application", { name: "Canvas viewport" })).toBeVisible();
  await expect(page.getByText("Duck Agent Cards", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Hero", exact: true })).toHaveAttribute(
    "aria-pressed",
    "true",
  );
  await expect(page.locator(".duck-card__image")).toHaveAttribute("src", "/img.jpg");
  await expect(page.getByRole("button", { name: "Export image" })).toBeVisible();
});

test("browser: canvas drop replaces default duck photo", async ({ page }) => {
  await page.goto("/");

  const upload = await page.evaluateHandle(() => {
    const dataTransfer = new DataTransfer();
    const file = new File(
      [
        '<svg xmlns="http://www.w3.org/2000/svg" width="128" height="96"><rect width="128" height="96" fill="#7c3aed"/></svg>',
      ],
      "replacement.svg",
      { type: "image/svg+xml" },
    );

    dataTransfer.items.add(file);
    return dataTransfer;
  });

  await page
    .getByRole("application", { name: "Canvas viewport" })
    .dispatchEvent("drop", { dataTransfer: upload });

  await expect(page.locator(".duck-card__image")).toHaveAttribute("src", /^data:image\/svg\+xml/);
});
