import { test } from "@playwright/test";

test("capture dashboard", async ({ page }) => {
  test.setTimeout(60000);

  // Login with demo user instead of creating new one
  await page.goto("/login");
  await page.fill('input[type="email"]', "demo@neobank.com");
  await page.fill('input[type="password"]', "password123");
  await page.click('button[type="submit"]');

  // Dash
  await page.waitForURL("/dashboard", { timeout: 15000 });
  // Wait for page to load without depending on account cards (may not exist)
  await page.waitForTimeout(3000);
  await page.screenshot({ path: "screenshots/dashboard.png", fullPage: true });
});
