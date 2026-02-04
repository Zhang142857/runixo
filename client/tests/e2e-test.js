const { chromium } = require('@playwright/test');
const path = require('path');
const fs = require('fs');
const SCREENSHOTS_DIR = path.join(__dirname, 'screenshots');
if (!fs.existsSync(SCREENSHOTS_DIR)) {
  fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
}
async function runTests() {
  const consoleErrors = [];
  const consoleWarnings = [];
  const testResults = [];
  console.log('============================================================');
  console.log('ServerHub Client Automation Test');
  console.log('============================================================');
  const browser = await chromium.launch({ 
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 }
  });
  const page = await context.newPage();
  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    } else if (msg.type() === 'warning') {
      consoleWarnings.push(msg.text());
    }
  });
  page.on('pageerror', error => {
    consoleErrors.push('Page Error: ' + error.message);
  });
  try {
    console.log('[Test 1] Navigate to homepage...');
    await page.goto('http://localhost:5173', { 
      waitUntil: 'networkidle',
      timeout: 30000 
    });
    console.log('  [PASS] Page loaded successfully');
    testResults.push({ name: 'Homepage Load', status: 'PASS' });
    await page.screenshot({ 
      path: path.join(SCREENSHOTS_DIR, '01-homepage.png'),
      fullPage: true 
    });
    console.log('  [PASS] Homepage screenshot saved');
    console.log('[Test 2] Check page structure...');
    const sidebar = await page.locator('.sidebar, [class*=sidebar], nav, aside').first();
    const sidebarVisible = await sidebar.isVisible().catch(() => false);
    if (sidebarVisible) {
      console.log('  [PASS] Sidebar exists');
      testResults.push({ name: 'Sidebar Check', status: 'PASS' });
    } else {
      console.log('  [FAIL] Sidebar not found');
      testResults.push({ name: 'Sidebar Check', status: 'FAIL' });
    }
    const header = await page.locator('header, .header, [class*=header]').first();
    const headerVisible = await header.isVisible().catch(() => false);
    if (headerVisible) {
      console.log('  [PASS] Header exists');
      testResults.push({ name: 'Header Check', status: 'PASS' });
    } else {
      console.log('  [FAIL] Header not found');
      testResults.push({ name: 'Header Check', status: 'FAIL' });
    }
    const mainContent = await page.locator('main, .main, [class*=content], [class*=main]').first();
    const mainVisible = await mainContent.isVisible().catch(() => false);
    if (mainVisible) {
      console.log('  [PASS] Main content area exists');
      testResults.push({ name: 'Main Content Check', status: 'PASS' });
    } else {
      console.log('  [FAIL] Main content area not found');
      testResults.push({ name: 'Main Content Check', status: 'FAIL' });
    }
    console.log('[Test 3] Test sidebar navigation...');
    const allLinks = await page.locator('a').all();
    console.log('  Found ' + allLinks.length + ' links on page');
    const menuKeywords = ['Server', 'Container', 'File', 'Terminal', 'Monitor', 'Setting', 'Dashboard', 'Home'];
    let screenshotIndex = 2;
    for (const keyword of menuKeywords) {
      try {
        const menuItem = page.locator('a, button').filter({ hasText: keyword }).first();
        const isVisible = await menuItem.isVisible().catch(() => false);
        if (isVisible) {
          console.log('  Clicking menu: ' + keyword + '...');
          await menuItem.click();
          await page.waitForTimeout(1500);
          const screenshotName = '0' + screenshotIndex + '-' + keyword.toLowerCase() + '.png';
          await page.screenshot({ 
            path: path.join(SCREENSHOTS_DIR, screenshotName),
            fullPage: true 
          });
          console.log('  [PASS] Screenshot saved: ' + screenshotName);
          testResults.push({ name: 'Navigate to ' + keyword, status: 'PASS' });
          screenshotIndex++;
        }
      } catch (e) {
        console.log('  [INFO] Menu ' + keyword + ' not found or not clickable');
      }
    }
  } catch (error) {
    console.log('[FAIL] Test error: ' + error.message);
    testResults.push({ name: 'Test Execution', status: 'FAIL', error: error.message });
    try {
      await page.screenshot({ 
        path: path.join(SCREENSHOTS_DIR, 'error-state.png'),
        fullPage: true 
      });
    } catch (e) {}
  }
  await browser.close();
  console.log('============================================================');
  console.log('Test Report');
  console.log('============================================================');
  console.log('[Test Results]');
  let passCount = 0;
  let failCount = 0;
  for (const result of testResults) {
    console.log('  ' + (result.status === 'PASS' ? '[PASS]' : '[FAIL]') + ' ' + result.name);
    if (result.status === 'PASS') passCount++;
    else failCount++;
  }
  console.log('  Total: ' + passCount + ' passed, ' + failCount + ' failed');
  console.log('[JavaScript Errors]');
  if (consoleErrors.length === 0) {
    console.log('  [PASS] No JavaScript errors found');
  } else {
    console.log('  Found ' + consoleErrors.length + ' errors:');
    consoleErrors.forEach((err, i) => console.log('  ' + (i+1) + '. ' + err));
  }
  console.log('[Console Warnings]');
  if (consoleWarnings.length === 0) {
    console.log('  [PASS] No console warnings found');
  } else {
    console.log('  Found ' + consoleWarnings.length + ' warnings');
  }
  console.log('[Screenshots]');
  const screenshots = fs.readdirSync(SCREENSHOTS_DIR).filter(f => f.endsWith('.png'));
  screenshots.forEach(s => console.log('  - ' + path.join(SCREENSHOTS_DIR, s)));
  console.log('============================================================');
  console.log('Test Complete');
  console.log('============================================================');
}
runTests().catch(console.error);
