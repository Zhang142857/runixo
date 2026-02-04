const { test, expect } = require('@playwright/test');
const path = require('path');

test.describe('ServerHub Pages Test', () => {
  test.beforeEach(async ({ page }) => {
    page.setDefaultTimeout(30000);
  });

  test('AI Assistant Page', async ({ page }) => {
    await page.goto('http://localhost:5173/#/ai');
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'screenshots/01-ai-page.png', 
      fullPage: true 
    });
  });

  test('Processes Page', async ({ page }) => {
    await page.goto('http://localhost:5173/#/processes');
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'screenshots/02-processes-page.png', 
      fullPage: true 
    });
  });

  test('Compose Page', async ({ page }) => {
    await page.goto('http://localhost:5173/#/compose');
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'screenshots/03-compose-page.png', 
      fullPage: true 
    });
  });

  test('Sidebar Navigation', async ({ page }) => {
    await page.goto('http://localhost:5173/#/');
    await page.waitForTimeout(2000);
    
    const html = await page.content();
    
    console.log('Checking menu items...');
    console.log('AI menu:', html.includes('AI') || html.includes('助手'));
    console.log('Process menu:', html.includes('进程') || html.includes('Process'));
    console.log('Compose menu:', html.includes('Compose') || html.includes('编排'));
    
    await page.screenshot({ 
      path: 'screenshots/04-sidebar.png', 
      fullPage: true 
    });
  });
});
