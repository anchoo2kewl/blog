#!/usr/bin/env node
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  page.setDefaultTimeout(5000);
  try {
    await page.goto('http://localhost:22222/admin/formatting-guide', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);
    const path = 'debug_formatting_light.png';
    await page.screenshot({ path, fullPage: true });
    console.log('Saved screenshot:', path);
  } finally {
    await browser.close();
  }
})();

