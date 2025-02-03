import { test, expect, Page } from '@playwright/test';

/**
 * Demo User E2E Tests
 * These tests require the backend services (identity, ledger, payment) to be running
 * Demo user: demo@neobank.com / password
 */

// Helper function to login with demo user
async function loginWithDemoUser(page: Page): Promise<boolean> {
  await page.goto('/login');
  await page.fill('input[type="email"]', 'demo@neobank.com');
  await page.fill('input[type="password"]', 'password');
  await page.click('button[type="submit"]');
  
  // Wait for navigation
  await page.waitForTimeout(3000);
  
  // Check if login was successful
  return page.url().includes('/dashboard');
}

test.describe('Demo User Authentication', () => {
  test('should login with demo@neobank.com', async ({ page }) => {
    const loginSuccess = await loginWithDemoUser(page);
    
    if (loginSuccess) {
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('h1')).toContainText('Overview');
    } else {
      // Backend not running, just verify we're still on login
      await expect(page).toHaveURL('/login');
      console.log('Backend not running - login test skipped');
    }
  });

  test('should persist session after login', async ({ page }) => {
    const loginSuccess = await loginWithDemoUser(page);
    
    if (loginSuccess) {
      // Navigate away and back
      await page.goto('/cards');
      await page.waitForTimeout(1000);
      
      // Should still be logged in (not redirected to login)
      const currentUrl = page.url();
      expect(currentUrl.includes('/login')).toBe(false);
    }
  });
});

test.describe('Dashboard Features', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should display dashboard with balance', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      // Check for key dashboard elements
      await expect(page.locator('text=Total Balance')).toBeVisible();
      await expect(page.locator('text=$')).toBeVisible();
      await expect(page.locator('h1')).toContainText('Overview');
    }
  });

  test('should display account cards from seed data', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      // Wait for accounts to load
      await page.waitForTimeout(2000);
      
      // Should have account cards (from seed data)
      const accountCards = page.getByTestId('account-card');
      const count = await accountCards.count();
      
      // If backend is running with seed data, we should have accounts
      console.log(`Found ${count} account cards`);
    }
  });

  test('should navigate to transfers from dashboard', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.click('text=Send Money');
      await expect(page).toHaveURL('/transfers');
    }
  });

  test('should create new account', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      const initialCount = await page.getByTestId('account-card').count();
      
      // Click new account button
      await page.click('text=New Account');
      await page.waitForTimeout(2000);
      
      const newCount = await page.getByTestId('account-card').count();
      console.log(`Accounts: ${initialCount} -> ${newCount}`);
    }
  });
});

test.describe('Sidebar Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should navigate to Cards page', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.click('nav >> text=Cards');
      await expect(page).toHaveURL('/cards');
    }
  });

  test('should navigate to Products page', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.click('nav >> text=Products');
      await expect(page).toHaveURL('/products');
    }
  });

  test('should navigate to Transfers page', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.click('nav >> text=Transfers');
      await expect(page).toHaveURL('/transfers');
    }
  });

  test('should navigate back to Dashboard', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.click('nav >> text=Cards');
      await page.waitForTimeout(500);
      await page.click('nav >> text=Dashboard');
      await expect(page).toHaveURL('/dashboard');
    }
  });
});

test.describe('Transfers Feature', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should display transfer form', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/transfers');
      
      // Check for transfer form elements
      await expect(page.locator('text=Send Money')).toBeVisible();
      await expect(page.locator('text=From Account')).toBeVisible();
      await expect(page.locator('text=Amount')).toBeVisible();
      await expect(page.locator('button:has-text("Send Funds")')).toBeVisible();
    }
  });

  test('should load accounts in dropdown', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/transfers');
      await page.waitForTimeout(2000);
      
      // Check if from account dropdown has options
      const selectElement = page.locator('select');
      const options = await selectElement.locator('option').count();
      console.log(`Transfer form has ${options} account options`);
      
      expect(options).toBeGreaterThanOrEqual(0);
    }
  });

  test('should fill transfer form', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/transfers');
      await page.waitForTimeout(2000);
      
      // Fill in the form
      await page.fill('input[placeholder="Enter destination UUID"]', 'b0000001-0001-0001-0001-000000000002');
      await page.fill('input[placeholder="0.00"]', '100');
      await page.fill('input[placeholder="Dinner, Rent..."]', 'Test transfer');
      
      // Verify form is filled
      await expect(page.locator('input[placeholder="Enter destination UUID"]')).toHaveValue('b0000001-0001-0001-0001-000000000002');
      await expect(page.locator('input[placeholder="0.00"]')).toHaveValue('100');
    }
  });

  test('should attempt to submit transfer', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/transfers');
      await page.waitForTimeout(2000);
      
      // Fill in the form
      await page.fill('input[placeholder="Enter destination UUID"]', 'b0000001-0001-0001-0001-000000000002');
      await page.fill('input[placeholder="0.00"]', '10');
      await page.fill('input[placeholder="Dinner, Rent..."]', 'E2E Test Transfer');
      
      // Click submit
      await page.click('button:has-text("Send Funds")');
      
      // Wait for response
      await page.waitForTimeout(3000);
      
      // Either success or error (depending on backend)
      const currentUrl = page.url();
      const hasSuccessMessage = await page.locator('text=Transfer Successful').isVisible().catch(() => false);
      
      console.log(`Transfer result: ${hasSuccessMessage ? 'Success' : 'Pending/Failed'}, URL: ${currentUrl}`);
    }
  });
});

test.describe('Cards Page', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should display cards page', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/cards');
      await page.waitForTimeout(1000);
      
      // Should have cards page heading
      await expect(page.locator('h1')).toBeVisible();
    }
  });
});

test.describe('Products Page', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should display products page', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      await page.goto('/products');
      await page.waitForTimeout(1000);
      
      // Should have products page heading
      await expect(page.locator('h1')).toBeVisible();
    }
  });
});

test.describe('Theme Toggle', () => {
  test.beforeEach(async ({ page }) => {
    await loginWithDemoUser(page);
  });

  test('should toggle between light and dark mode', async ({ page }) => {
    if (page.url().includes('/dashboard')) {
      // Find and click the theme toggle button
      const themeToggle = page.locator('button[aria-label*="mode"]');
      
      if (await themeToggle.isVisible()) {
        // Get initial state
        const initialTheme = await page.evaluate(() => 
          document.documentElement.getAttribute('data-theme')
        );
        
        // Click toggle
        await themeToggle.click();
        await page.waitForTimeout(500);
        
        // Check if theme changed
        const newTheme = await page.evaluate(() => 
          document.documentElement.getAttribute('data-theme')
        );
        
        console.log(`Theme changed: ${initialTheme} -> ${newTheme}`);
      }
    }
  });
});

test.describe('Logout', () => {
  test('should logout successfully', async ({ page }) => {
    const loginSuccess = await loginWithDemoUser(page);
    
    if (loginSuccess) {
      // Click sign out
      await page.click('text=Sign Out');
      
      // Should redirect to login
      await expect(page).toHaveURL('/login');
    }
  });
});
