import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Visual Screenshots', () => {
    // Ensure screenshots directory exists (Playwright creates it automatically usually)
    
    test('capture landing, login, and dashboard', async ({ page }) => {
        // 1. Landing Page
        await page.goto('/');
        await page.waitForTimeout(1000); // Allow animations to settle
        await page.screenshot({ path: 'screenshots/landing.png', fullPage: true });

        // 2. Login Page
        await page.goto('/login');
        await page.waitForTimeout(1000);
        await page.screenshot({ path: 'screenshots/login.png', fullPage: true });

        // 3. Register Page
        await page.goto('/register');
        await page.waitForTimeout(1000);
        await page.screenshot({ path: 'screenshots/register.png', fullPage: true });

        // 4. Dashboard (Login flow)
        await page.goto('/login');
        await page.fill('input[type="email"]', 'demo@example.com'); // This user might not exist if DB clean
        await page.fill('input[type="password"]', 'password');
        
        // We'll try to register a new one to be safe for the screenshot
        await page.goto('/register');
        const uniqueEmail = `visual_${Date.now()}@example.com`;
        await page.fill('input[name="firstName"]', "Visual");
        await page.fill('input[name="lastName"]', "User");
        await page.fill('input[type="email"]', uniqueEmail);
        await page.fill('input[type="password"]', "password123");
        await page.click('button[type="submit"]'); // Get Started
        
        // Should redirect to login? Or auto login?
        // App logic: router.push("/login?registered=true");
        // So we need to log in now.
        await page.waitForURL(/\/login/);
        await page.fill('input[type="email"]', uniqueEmail);
        await page.fill('input[type="password"]', "password123");
        await page.click('button[type="submit"]');
        
        await page.waitForURL('/dashboard');
        await page.waitForSelector('[data-testid="account-card"]'); // Wait for accounts to load
        await page.waitForTimeout(2000); // Wait for chart animation
        await page.screenshot({ path: 'screenshots/dashboard.png', fullPage: true });
    });
});
