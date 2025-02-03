import { test, expect } from '@playwright/test';

test('capture dashboard', async ({ page }) => {
    test.setTimeout(60000);
    // Register unique user
    await page.goto('/register');
    const uniqueEmail = `dash_${Date.now()}@example.com`;
    await page.fill('input[name="firstName"]', "Dash");
    await page.fill('input[name="lastName"]', "User");
    await page.fill('input[type="email"]', uniqueEmail);
    await page.fill('input[type="password"]', "password123");
    await page.click('button[type="submit"]');
    
    // Login
    await page.waitForURL(/\/login/);
    await page.fill('input[type="email"]', uniqueEmail);
    await page.fill('input[type="password"]', "password123");
    await page.click('button[type="submit"]');
    
    // Dash
    await page.waitForURL('/dashboard');
    await page.waitForSelector('[data-testid="account-card"]', { timeout: 30000 });
    await page.waitForTimeout(2000);
    await page.screenshot({ path: 'screenshots/dashboard.png', fullPage: true });
});
