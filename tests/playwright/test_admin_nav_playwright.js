#!/usr/bin/env node

const { chromium } = require('playwright');

async function testAdminNavigation() {
    let browser;
    try {
        console.log('🚀 Starting Playwright browser...');
        browser = await chromium.launch({ headless: true });
        const page = await browser.newPage();
        // Fast-fail defaults
        page.setDefaultTimeout(5000);
        page.setDefaultNavigationTimeout(5000);

        console.log('🔍 Testing admin navigation...');

        // Step 1: Go to signin page
        console.log('1️⃣ Navigating to signin page...');
        await page.goto('http://localhost:22222/signin', { waitUntil: 'domcontentloaded' });
        await page.waitForLoadState('networkidle', { timeout: 1500 }).catch(() => {});

        // Step 2: Fill login form
        console.log('2️⃣ Filling login form with admin credentials...');
        await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
        await page.fill('input[name="password"]', '123456qwertyu');

        // Step 3: Submit form
        console.log('3️⃣ Submitting login form...');
        await page.click('button[type="submit"]');
        await page.waitForNavigation({ waitUntil: 'domcontentloaded', timeout: 2000 }).catch(() => {
            console.log('   - Navigation timeout, checking current page...');
        });

        // Step 4: Check current URL and page content
        console.log('4️⃣ Current URL after login:', page.url());

        // Wait a bit for any dynamic content
        await page.waitForTimeout(500);

        // Step 5: Look for navigation elements
        console.log('5️⃣ Checking for navigation elements...');

        const userDropdown = await page.locator('.user-dropdown').count();
        const userButton = await page.locator('.user-dropdown-btn').count();
        const signOutIcon = await page.locator('a[aria-label="Sign Out"]:visible').count();
        const adminTab = await page.locator('a[href="/admin/posts"].admin-link:visible').count();
        const profileLinkInMenu = await page.locator('.user-dropdown-menu a[href="/users/me"]').count();
        const apiAccessLinkInMenu = await page.locator('.user-dropdown-menu a[href="/api-access"]').count();

        console.log('📊 Navigation Elements Found:');
        console.log(`   - User dropdown: ${userDropdown > 0 ? '✅ YES (' + userDropdown + ')' : '❌ NO'}`);
        console.log(`   - User button: ${userButton > 0 ? '✅ YES (' + userButton + ')' : '❌ NO'}`);
        console.log(`   - Admin tab visible: ${adminTab > 0 ? '✅ YES' : '❌ NO'}`);
        console.log(`   - Sign Out icon: ${signOutIcon > 0 ? '✅ YES' : '❌ NO'}`);
        console.log(`   - Profile in menu: ${profileLinkInMenu > 0 ? '✅ YES' : '❌ NO'}`);
        console.log(`   - API Access in menu: ${apiAccessLinkInMenu > 0 ? '✅ YES' : '❌ NO'}`);

        // Screenshot for debugging
        console.log('\n📸 Taking screenshot...');
        await page.screenshot({ path: 'debug_nav.png', fullPage: true });
        console.log('   - Screenshot saved to: debug_nav.png');

    } catch (error) {
        console.error('❌ Error:', error.message);
        console.error('Stack:', error.stack);
        process.exit(1);
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

testAdminNavigation();

