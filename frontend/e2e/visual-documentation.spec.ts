import { test, expect, Page } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Comprehensive Visual Documentation Tests
 * Captures screenshots of all major features and flows
 */

const SCREENSHOT_DIR = 'screenshots/comprehensive';

// Demo account UUIDs from seed data
const DEMO_ACCOUNTS = {
  highYieldSavings: 'b0000001-0001-0001-0001-000000000002',
  johnDoeChecking: 'b0000002-0002-0002-0002-000000000001',
};

function ensureScreenshotDir() {
  const dir = path.join(process.cwd(), SCREENSHOT_DIR);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

async function screenshot(page: Page, name: string) {
  ensureScreenshotDir();
  await page.screenshot({ 
    path: `${SCREENSHOT_DIR}/${name}.png`, 
    fullPage: true 
  });
  console.log(`Screenshot saved: ${name}.png`);
}

async function loginWithDemoUser(page: Page): Promise<boolean> {
  await page.goto('/login');
  await page.waitForTimeout(500);
  await page.fill('input[type="email"]', 'demo@neobank.com');
  await page.fill('input[type="password"]', 'password');
  await page.click('button[type="submit"]');
  
  // Wait for navigation with timeout
  try {
    await page.waitForURL('**/dashboard', { timeout: 10000 });
    return true;
  } catch {
    return page.url().includes('/dashboard');
  }
}

// ============= PUBLIC PAGES =============
test.describe('01 - Public Pages', () => {
  test('capture landing page', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1500);
    await screenshot(page, '01_landing_page');
  });

  test('capture login page', async ({ page }) => {
    await page.goto('/login');
    await page.waitForTimeout(1000);
    await screenshot(page, '02_login_page');
  });

  test('capture register page', async ({ page }) => {
    await page.goto('/register');
    await page.waitForTimeout(1000);
    await screenshot(page, '03_register_page');
  });
});

// ============= DASHBOARD =============
test.describe('02 - Dashboard', () => {
  test('capture dashboard overview', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.waitForTimeout(2000);
    await screenshot(page, '04_dashboard_overview');
  });

  test('capture dashboard with accounts', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    // Wait for accounts to load
    try {
      await page.waitForSelector('[data-testid="account-card"]', { timeout: 15000 });
    } catch {
      console.log('No account cards found, capturing anyway');
    }
    await page.waitForTimeout(1000);
    await screenshot(page, '05_dashboard_with_accounts');
  });
});

// ============= CARDS PAGE =============
test.describe('03 - Cards Page', () => {
  test('capture cards page', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.click('nav >> text=Cards');
    await page.waitForURL('**/cards', { timeout: 5000 });
    await page.waitForTimeout(2000);
    await screenshot(page, '06_cards_page');
  });
});

// ============= PRODUCTS PAGE =============
test.describe('04 - Products Page', () => {
  test('capture products page', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.click('nav >> text=Products');
    await page.waitForURL('**/products', { timeout: 10000 });
    // Wait for product cards or content to load
    try {
      await page.waitForSelector('[class*="product"], [class*="card"], h1, h2', { timeout: 5000 });
    } catch {
      console.log('No product elements found, capturing anyway');
    }
    await page.waitForTimeout(3000);
    await screenshot(page, '07_products_page');
  });
});

// ============= TRANSFERS FLOW =============
test.describe('05 - Transfer Flow', () => {
  test('capture transfer form empty', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.goto('/transfers');
    await page.waitForTimeout(2500);
    await screenshot(page, '08_transfer_form_empty');
  });

  test('capture transfer form filled - own account', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.goto('/transfers');
    await page.waitForTimeout(2000);
    
    // Fill the form for own account transfer
    await page.fill('input[placeholder="Enter destination UUID"]', DEMO_ACCOUNTS.highYieldSavings);
    // Use the large amount input field
    await page.locator('input[type="number"][placeholder="0.00"]').first().fill('100');
    await page.fill('input[placeholder*="What\'s this for"]', 'Monthly savings deposit');
    
    await page.waitForTimeout(500);
    await screenshot(page, '09_transfer_to_own_account');
  });

  test('capture transfer form filled - other user', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.goto('/transfers');
    await page.waitForTimeout(2000);
    
    // Fill form for transfer to John Doe
    await page.fill('input[placeholder="Enter destination UUID"]', DEMO_ACCOUNTS.johnDoeChecking);
    await page.locator('input[type="number"][placeholder="0.00"]').first().fill('50');
    await page.fill('input[placeholder*="What\'s this for"]', 'Payment to John Doe');
    
    await page.waitForTimeout(500);
    await screenshot(page, '10_transfer_to_other_user');
  });

  test('capture transfer submission and result', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.goto('/transfers');
    await page.waitForTimeout(2500);
    
    await page.fill('input[placeholder="Enter destination UUID"]', DEMO_ACCOUNTS.highYieldSavings);
    await page.locator('input[type="number"][placeholder="0.00"]').first().fill('25');
    await page.fill('input[placeholder*="What\'s this for"]', 'Screenshot test transfer');
    
    // Screenshot before submit (filled form)
    await screenshot(page, '11_transfer_filled_ready');
    
    // Click Review Transfer to open confirmation modal
    await page.click('button:has-text("Review Transfer")');
    await page.waitForTimeout(500);
    
    // Screenshot the confirmation modal
    await screenshot(page, '11b_transfer_confirm_modal');
    
    // Click Confirm & Send
    await page.click('button:has-text("Confirm & Send")');
    
    // Wait for success message, redirect, or error
    await page.waitForTimeout(5000);
    await screenshot(page, '12_transfer_result');
  });
});

// ============= THEME TOGGLE =============
test.describe('06 - Theme Toggle', () => {
  test('capture dark mode dashboard', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.waitForTimeout(1500);
    await screenshot(page, '13_theme_dark_mode');
  });

  test('capture light mode dashboard', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.waitForTimeout(1000);
    
    // Find and click the theme toggle
    const themeToggle = page.locator('button[aria-label*="mode"], button[aria-label*="theme"]');
    if (await themeToggle.count() > 0) {
      await themeToggle.first().click();
      await page.waitForTimeout(500);
    } else {
      // Try clicking by looking for the toggle component
      const toggleArea = page.locator('text=Dark Mode, text=Light Mode').first();
      if (await toggleArea.isVisible()) {
        await toggleArea.click();
        await page.waitForTimeout(500);
      }
    }
    
    await screenshot(page, '14_theme_light_mode');
  });
});

// ============= SIDEBAR =============
test.describe('07 - Sidebar Navigation', () => {
  test('capture sidebar', async ({ page }) => {
    const loggedIn = await loginWithDemoUser(page);
    expect(loggedIn).toBe(true);
    
    await page.waitForTimeout(1000);
    
    // Capture full page which includes sidebar
    await screenshot(page, '15_sidebar_full');
    
    // Also capture just the sidebar
    const sidebar = page.locator('aside');
    if (await sidebar.isVisible()) {
      await sidebar.screenshot({ path: `${SCREENSHOT_DIR}/15b_sidebar_only.png` });
    }
  });
});

// ============= RESPONSIVE / MOBILE =============
test.describe('08 - Mobile Responsive', () => {
  test('capture mobile landing page', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 932 }); // iPhone 14 Pro Max
    await page.goto('/');
    await page.waitForTimeout(2000);
    await screenshot(page, '16_mobile_landing');
  });

  test('capture mobile login page', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 932 });
    await page.goto('/login');
    await page.waitForTimeout(1500);
    await screenshot(page, '17_mobile_login');
  });

  test('capture mobile register page', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 932 });
    await page.goto('/register');
    await page.waitForTimeout(1500);
    await screenshot(page, '18_mobile_register');
  });

  test('capture mobile dashboard', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 932 });
    const loggedIn = await loginWithDemoUser(page);
    if (loggedIn) {
      await page.waitForTimeout(2000);
      await screenshot(page, '19_mobile_dashboard');
    }
  });
});
