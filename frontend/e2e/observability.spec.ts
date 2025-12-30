import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

const screenshotDir = path.join(__dirname, '..', 'observability_screenshots');

// Ensure screenshot directory exists
if (!fs.existsSync(screenshotDir)) {
  fs.mkdirSync(screenshotDir);
}

test('Capture Observability Dashboards', async ({ page }) => {
  test.setTimeout(120000); // 2 minutes

  // 1. Prometheus
  /*
  try {
    console.log('Navigating to Prometheus...');
    await page.goto('http://localhost:9090/graph?g0.expr=http_requests_total&g0.tab=0&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h');
    await page.waitForTimeout(5000); 
    await page.screenshot({ path: path.join(screenshotDir, 'prometheus.png'), fullPage: true });
    console.log('Captured Prometheus');
  } catch (e) {
    console.error('Failed to capture Prometheus', e);
  }

  */

  // 2. Jaeger
  try {
    console.log('Navigating to Jaeger...');
    await page.goto('http://localhost:16686/search?service=payment-service&limit=20&lookback=1h', { timeout: 30000 });
    await page.waitForTimeout(5000);
    // Reload to ensure traces appear?
    await page.reload();
    await page.waitForTimeout(5000);

    await page.screenshot({ path: path.join(screenshotDir, 'jaeger_traces.png'), fullPage: true });
    console.log('Captured Jaeger');
  } catch (e) {
    console.error('Failed to capture Jaeger', e);
  }


  // 3. Grafana
  try {
    console.log('Navigating to Grafana...');
    await page.goto('http://localhost:3000/login', { timeout: 30000 });

    await page.fill('input[name="user"]', 'admin');
    await page.fill('input[name="password"]', 'admin');
    
    try {
        await page.getByRole('button', { name: 'Log in' }).click({ timeout: 5000 });
    } catch {
        await page.click('button[type="submit"]');
    }
    
    // Handle "New password" prompt
    try {
      const skipButton = page.getByText('Skip');
      await skipButton.waitFor({ state: 'visible', timeout: 5000 });
      await skipButton.click();
    } catch {}

    await page.waitForTimeout(2000);

    // Navigate directly to NeoBank Dashboard
    const dashboardUrl = 'http://localhost:3000/d/neobank-overview/neobank-services-overview';
    console.log(`Navigating to ${dashboardUrl}...`);
    await page.goto(dashboardUrl, { timeout: 30000 });

    // Wait for dashboard to load (panels)
    await page.waitForTimeout(10000); // Increased wait for panels
    
    // Screenshot
    await page.screenshot({ path: path.join(screenshotDir, 'grafana_dashboard_full.png'), fullPage: true });
    console.log('Captured Grafana Dashboard');
  } catch (e) {
    console.error('Failed to capture Grafana', e);
    await page.screenshot({ path: path.join(screenshotDir, 'grafana_error.png') });
  }
});
