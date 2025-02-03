import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  const email = `testuser_${Date.now()}@example.com`;
  const password = 'password123';

  test('should allow a user to register and login', async ({ page }) => {
    // 1. Register
    await page.goto('/login');
    await page.click('text=Create an account');
    await expect(page).toHaveURL('/register');

    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    await page.fill('input[placeholder="First Name"]', 'Test'); // Assuming placeholders match
    await page.fill('input[placeholder="Last Name"]', 'User');
    
    // Check if there is a submit button with correct text
    await page.click('button[type="submit"]');

    // Expect redirect to login or dashboard
    await expect(page).toHaveURL(/\/login/);

    // 2. Login
    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    await page.click('button[type="submit"]');

    // Expect dashboard
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('h1')).toContainText('Overview');
  });
});
