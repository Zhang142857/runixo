import { chromium } from 'playwright';
import path from 'path';
import fs from 'fs';

async function testCloudflareIntegration() {
  const screenshotDir = 'C:/Users/Administrator/Desktop/serverhub/screenshots';
  
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }

  console.log('启动浏览器...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await context.newPage();

  try {
    console.log('1. 访问首页...');
    await page.goto('http://localhost:5177/#/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(screenshotDir, '01-homepage.png'), fullPage: true });
    console.log('   首页截图已保存');

    console.log('2. 导航到云服务页面...');
    await page.goto('http://localhost:5177/#/cloud');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(screenshotDir, '02-cloud-services.png'), fullPage: true });
    console.log('   云服务页面截图已保存');

    console.log('3. 导航到 Cloudflare 管理页面...');
    await page.goto('http://localhost:5177/#/cloud/cloudflare');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1500);
    await page.screenshot({ path: path.join(screenshotDir, '03-cloudflare-dns.png'), fullPage: true });
    console.log('   Cloudflare DNS 标签页截图已保存 (默认标签)');

    // 检查页面是否正确加载
    const pageTitle = await page.locator('h1').first().innerText().catch(() => '');
    console.log('   页面标题:', pageTitle);

    // Element Plus tabs 使用 .el-tabs__item 类
    console.log('4. 测试 SSL/TLS 标签页...');
    try {
      const sslTab = page.locator('.el-tabs__item:has-text("SSL/TLS")');
      if (await sslTab.isVisible({ timeout: 3000 })) {
        await sslTab.click();
        await page.waitForTimeout(800);
        await page.screenshot({ path: path.join(screenshotDir, '04-cloudflare-ssl.png'), fullPage: true });
        console.log('   SSL/TLS 标签页截图已保存');
      } else {
        console.log('   SSL/TLS 标签不可见');
      }
    } catch (e) {
      console.log('   SSL/TLS 标签操作失败:', e.message);
    }

    console.log('5. 测试缓存标签页...');
    try {
      const cacheTab = page.locator('.el-tabs__item:has-text("缓存")');
      if (await cacheTab.isVisible({ timeout: 3000 })) {
        await cacheTab.click();
        await page.waitForTimeout(800);
        await page.screenshot({ path: path.join(screenshotDir, '05-cloudflare-cache.png'), fullPage: true });
        console.log('   缓存标签页截图已保存');
      } else {
        console.log('   缓存标签不可见');
      }
    } catch (e) {
      console.log('   缓存标签操作失败:', e.message);
    }

    console.log('6. 测试 Tunnel 标签页...');
    try {
      const tunnelTab = page.locator('.el-tabs__item:has-text("Tunnel")');
      if (await tunnelTab.isVisible({ timeout: 3000 })) {
        await tunnelTab.click();
        await page.waitForTimeout(800);
        await page.screenshot({ path: path.join(screenshotDir, '06-cloudflare-tunnel.png'), fullPage: true });
        console.log('   Tunnel 标签页截图已保存');
      } else {
        console.log('   Tunnel 标签不可见');
      }
    } catch (e) {
      console.log('   Tunnel 标签操作失败:', e.message);
    }

    // 返回 DNS 标签页测试添加记录功能
    console.log('7. 测试 DNS 添加记录对话框...');
    try {
      const dnsTab = page.locator('.el-tabs__item:has-text("DNS")');
      if (await dnsTab.isVisible({ timeout: 3000 })) {
        await dnsTab.click();
        await page.waitForTimeout(500);
      }
      
      const addButton = page.locator('button:has-text("添加记录")');
      if (await addButton.isVisible({ timeout: 3000 })) {
        await addButton.click();
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(screenshotDir, '07-cloudflare-add-dns-dialog.png'), fullPage: true });
        console.log('   DNS 添加记录对话框截图已保存');
        
        // 关闭对话框
        const cancelBtn = page.locator('.el-dialog button:has-text("取消")');
        if (await cancelBtn.isVisible({ timeout: 2000 })) {
          await cancelBtn.click();
          await page.waitForTimeout(300);
        }
      } else {
        console.log('   添加记录按钮不可见');
      }
    } catch (e) {
      console.log('   DNS 添加记录对话框操作失败:', e.message);
    }

    console.log('\n========================================');
    console.log('测试完成！所有截图已保存到:', screenshotDir);
    console.log('========================================');

  } catch (error) {
    console.error('测试过程中出错:', error.message);
    await page.screenshot({ path: path.join(screenshotDir, 'error-screenshot.png'), fullPage: true });
  } finally {
    await browser.close();
    console.log('浏览器已关闭');
  }
}

testCloudflareIntegration();
