const { chromium } = require('playwright');

const screenshotDir = 'C:/Users/Administrator/Desktop/serverhub/screenshots';

async function runTests() {
  console.log('启动浏览器...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 }
  });
  const page = await context.newPage();

  const consoleLogs = [];
  page.on('console', msg => {
    consoleLogs.push({ type: msg.type(), text: msg.text() });
  });

  try {
    console.log('\n=== 测试仪表盘页面 ===');
    await page.goto('http://localhost:5173/#/', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: screenshotDir + '/01-dashboard.png',
      fullPage: true 
    });
    console.log('已保存: 01-dashboard.png');

    console.log('尝试点击告警卡片...');
    const alertCard = await page.locator('.el-card').filter({ hasText: '告警' }).first();
    if (await alertCard.isVisible().catch(() => false)) {
      await alertCard.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ 
        path: screenshotDir + '/02-alert-dialog.png',
        fullPage: true 
      });
      console.log('已保存: 02-alert-dialog.png');
      await page.keyboard.press('Escape');
      await page.waitForTimeout(500);
    }

    console.log('\n=== 测试设置页面 ===');
    await page.goto('http://localhost:5173/#/settings', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: screenshotDir + '/03-settings-general.png',
      fullPage: true 
    });
    console.log('已保存: 03-settings-general.png');

    const tabs = await page.locator('.el-tabs__item').all();
    console.log('找到 ' + tabs.length + ' 个标签');

    console.log('切换到 AI 助手标签...');
    const aiTab = await page.locator('.el-tabs__item').filter({ hasText: 'AI' }).first();
    if (await aiTab.isVisible().catch(() => false)) {
      await aiTab.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ 
        path: screenshotDir + '/04-settings-ai.png',
        fullPage: true 
      });
      console.log('已保存: 04-settings-ai.png');
    }

    console.log('切换到数据管理标签...');
    const dataTab = await page.locator('.el-tabs__item').filter({ hasText: '数据' }).first();
    if (await dataTab.isVisible().catch(() => false)) {
      await dataTab.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ 
        path: screenshotDir + '/05-settings-data.png',
        fullPage: true 
      });
      console.log('已保存: 05-settings-data.png');
    }

    console.log('切换到关于标签...');
    const aboutTab = await page.locator('.el-tabs__item').filter({ hasText: '关于' }).first();
    if (await aboutTab.isVisible().catch(() => false)) {
      await aboutTab.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ 
        path: screenshotDir + '/06-settings-about.png',
        fullPage: true 
      });
      console.log('已保存: 06-settings-about.png');
    }

    console.log('\n=== 浏览器控制台日志 ===');
    const errors = consoleLogs.filter(log => log.type === 'error');
    if (errors.length > 0) {
      console.log('错误数量: ' + errors.length);
    } else {
      console.log('无错误');
    }

    console.log('\n=== 测试完成 ===');

  } catch (error) {
    console.error('测试出错:', error.message);
    await page.screenshot({ 
      path: screenshotDir + '/error-screenshot.png',
      fullPage: true 
    });
  } finally {
    await browser.close();
  }
}

runTests();
