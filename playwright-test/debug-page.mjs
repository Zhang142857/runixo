import { chromium } from 'playwright';

async function debugPage() {
  console.log('启动浏览器...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();

  try {
    console.log('访问 Cloudflare 页面...');
    await page.goto('http://localhost:5177/cloud/cloudflare');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    const title = await page.title();
    console.log('页面标题:', title);

    const url = page.url();
    console.log('当前 URL:', url);

    const bodyText = await page.locator('body').innerText();
    console.log('\n页面文本内容:');
    console.log('-------------------');
    console.log(bodyText.substring(0, 2000));
    console.log('-------------------');

    console.log('\n查找 Element Plus tabs 元素...');
    const elTabs = await page.locator('.el-tabs').count();
    console.log('el-tabs 元素数量:', elTabs);
    
    const elTabItems = await page.locator('.el-tabs__item').all();
    console.log('el-tabs__item 元素数量:', elTabItems.length);
    for (const item of elTabItems) {
      const text = await item.innerText().catch(() => '');
      console.log('  标签:', text);
    }

    const buttons = await page.locator('button').all();
    console.log('\n按钮数量:', buttons.length);

  } catch (error) {
    console.error('调试过程中出错:', error.message);
  } finally {
    await browser.close();
    console.log('\n浏览器已关闭');
  }
}

debugPage();
