#!/usr/bin/env node

const puppeteer = require('puppeteer');

async function testAdminNavigation() {
    let browser;
    try {
        browser = await puppeteer.launch({ headless: true });
        const page = await browser.newPage();
        
        console.log('🔍 Testing admin navigation...');
        
        // Step 1: Go to signin page
        console.log('1️⃣ Navigating to signin page...');
        await page.goto('http://localhost:22222/signin');
        
        // Step 2: Fill login form
        console.log('2️⃣ Filling login form...');
        await page.type('input[name="username"]', 'anchoo2kewl');
        await page.type('input[name="password"]', '123456qwertyu');
        
        // Step 3: Submit form
        console.log('3️⃣ Submitting login form...');
        await page.click('button[type="submit"]');
        await page.waitForNavigation({ waitUntil: 'networkidle2' });
        
        // Step 4: Check current URL
        console.log('4️⃣ Current URL:', page.url());
        
        // Step 5: Look for navigation elements
        console.log('5️⃣ Checking for navigation elements...');
        
        // Check for admin dropdown
        const adminDropdown = await page.$('.admin-dropdown');
        const adminButton = await page.$('.admin-dropdown-btn');
        const userDropdown = await page.$('.user-dropdown');
        const userButton = await page.$('.user-dropdown-btn');
        const signInLink = await page.$('a[href="/signin"]');
        
        console.log('📊 Navigation Elements Found:');
        console.log('   - Admin dropdown:', adminDropdown ? '✅ YES' : '❌ NO');
        console.log('   - Admin button:', adminButton ? '✅ YES' : '❌ NO');
        console.log('   - User dropdown:', userDropdown ? '✅ YES' : '❌ NO');
        console.log('   - User button:', userButton ? '✅ YES' : '❌ NO');
        console.log('   - Sign In link:', signInLink ? '✅ YES (should be NO)' : '❌ NO (correct)');
        
        // Get the actual navigation HTML
        const navHTML = await page.$eval('nav', el => el.outerHTML);
        console.log('\n📄 Navigation HTML (first 1000 chars):');
        console.log(navHTML.substring(0, 1000) + '...');
        
        // Check for specific text content
        const pageContent = await page.content();
        console.log('\n🔍 Text Content Search:');
        console.log('   - Contains "Admin":', pageContent.includes('Admin') ? '✅ YES' : '❌ NO');
        console.log('   - Contains "anchoo2kewl":', pageContent.includes('anchoo2kewl') ? '✅ YES' : '❌ NO');
        console.log('   - Contains "Sign In":', pageContent.includes('Sign In') ? '✅ YES' : '❌ NO');
        console.log('   - Contains "Sign Out":', pageContent.includes('Sign Out') ? '✅ YES' : '❌ NO');
        
    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

testAdminNavigation();