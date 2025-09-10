const { test, expect } = require('@playwright/test');

test.describe('Real Editor Page Layout Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/signin');
    await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
    await page.fill('input[name="password"]', '123456qwertyu');
    await page.click('button[type="submit"]');
    
    // Wait for redirect - it might go to home page, not admin
    await page.waitForLoadState('networkidle');
    
    // Navigate directly to the editor page
    await page.goto('/admin/posts/8/edit');
    await page.waitForLoadState('networkidle');
  });

  test('should have proper layout on actual editor page', async ({ page }) => {
    // Set desktop viewport
    await page.setViewportSize({ width: 1200, height: 800 });
    
    // Collect console messages
    const consoleLogs = [];
    page.on('console', msg => consoleLogs.push(msg.text()));
    
    // Wait for layout to settle and JavaScript to execute
    await page.waitForTimeout(3000);
    
    console.log('Current URL:', page.url());
    console.log('Console messages:', consoleLogs.filter(log => log.includes('Layout') || log.includes('Forcing')));
    
    // DIAGNOSTIC: Check the actual CSS being applied
    const layoutElement = page.locator('.editor-layout').first();
    const mainElement = page.locator('.editor-main').first();
    const sidebarElement = page.locator('.editor-sidebar').first();
    
    if (await layoutElement.isVisible().catch(() => false)) {
      const layoutStyles = await layoutElement.evaluate(el => {
        const styles = getComputedStyle(el);
        return {
          display: styles.display,
          flexDirection: styles.flexDirection,
          position: styles.position,
          width: styles.width,
          height: styles.height
        };
      });
      console.log('Layout element styles:', layoutStyles);
      
      if (await mainElement.isVisible().catch(() => false)) {
        const mainStyles = await mainElement.evaluate(el => {
          const styles = getComputedStyle(el);
          return {
            display: styles.display,
            flex: styles.flex,
            position: styles.position,
            float: styles.float,
            width: styles.width,
            order: styles.order
          };
        });
        console.log('Main element styles:', mainStyles);
      }
      
      if (await sidebarElement.isVisible().catch(() => false)) {
        const sidebarStyles = await sidebarElement.evaluate(el => {
          const styles = getComputedStyle(el);
          return {
            display: styles.display,
            flex: styles.flex,
            position: styles.position,
            float: styles.float,
            width: styles.width,
            order: styles.order,
            top: styles.top
          };
        });
        console.log('Sidebar element styles:', sidebarStyles);
        
        // Check actual positions
        const mainBox = await mainElement.boundingBox();
        const sidebarBox = await sidebarElement.boundingBox();
        
        console.log('Main box:', mainBox);
        console.log('Sidebar box:', sidebarBox);
        
        if (mainBox && sidebarBox) {
          const isStacked = sidebarBox.y > mainBox.y + 500;
          const isSideBySide = sidebarBox.x > mainBox.x + mainBox.width - 100;
          
          console.log('Layout analysis:');
          console.log('- Side by side:', isSideBySide);
          console.log('- Stacked (bad):', isStacked);
          
          if (isStacked) {
            console.log('❌ LAYOUT ISSUE: CSS is not working - sidebar pushed to bottom');
          } else if (isSideBySide) {
            console.log('✅ Layout fixed - sidebar is properly positioned to the right');
          } else {
            console.log('⚠️ Partial fix - sidebar positioning improved but not perfect');
          }
        }
      }
    } else {
      console.log('❌ Layout elements not found - check HTML structure');
    }
  });

  test('should show content length and check for overflow', async ({ page }) => {
    console.log('Current URL:', page.url());
    
    // Try to find content in different ways
    let contentLength = 0;
    
    const contentSelectors = ['textarea[name="content"]', '.wysiwyg', '#editor', 'textarea'];
    for (const selector of contentSelectors) {
      try {
        const element = page.locator(selector).first();
        if (await element.isVisible().catch(() => false)) {
          const content = await element.textContent();
          if (content && content.length > contentLength) {
            contentLength = content.length;
            console.log(`Found content via ${selector}: ${contentLength} characters`);
          }
        }
      } catch (e) {
        // Skip if selector fails
      }
    }
    
    // Check page dimensions
    const pageHeight = await page.evaluate(() => document.documentElement.scrollHeight);
    const viewportHeight = await page.evaluate(() => window.innerHeight);
    
    console.log(`Page height: ${pageHeight}px, Viewport height: ${viewportHeight}px`);
    console.log(`Total content length: ${contentLength} characters`);
    
    if (pageHeight > viewportHeight * 3) {
      console.log('⚠️ Very long content detected - this might cause layout issues');
    }
    
    if (contentLength > 10000) {
      console.log('✅ This appears to be the long test post');
    }
  });

  test('should check if sidebar elements are accessible after scrolling', async ({ page }) => {
    // Set desktop viewport
    await page.setViewportSize({ width: 1200, height: 800 });
    
    console.log('Current URL before scroll test:', page.url());
    
    // Find any sidebar-like elements
    const sidebarSelectors = [
      'input[name="slug"]',
      'input[name="featured_image_url"]',
      'button[type="submit"]',
      'input[type="submit"]',
      '.editor-sidebar',
      '.sidebar'
    ];
    
    const foundElements = [];
    for (const selector of sidebarSelectors) {
      try {
        const element = page.locator(selector).first();
        if (await element.isVisible().catch(() => false)) {
          foundElements.push(selector);
        }
      } catch (e) {
        // Skip if selector fails
      }
    }
    
    console.log('Found sidebar elements:', foundElements);
    
    if (foundElements.length > 0) {
      // Get initial positions
      const initialPositions = [];
      for (const selector of foundElements) {
        const element = page.locator(selector).first();
        const box = await element.boundingBox().catch(() => null);
        if (box) {
          initialPositions.push({ selector, y: box.y });
        }
      }
      
      console.log('Initial positions:', initialPositions);
      
      // Scroll down significantly
      await page.evaluate(() => window.scrollTo(0, 2000));
      await page.waitForTimeout(500);
      
      // Check positions after scroll
      const afterScrollPositions = [];
      for (const selector of foundElements) {
        const element = page.locator(selector).first();
        const visible = await element.isVisible().catch(() => false);
        const box = await element.boundingBox().catch(() => null);
        afterScrollPositions.push({ 
          selector, 
          visible, 
          y: box ? box.y : null 
        });
      }
      
      console.log('After scroll positions:', afterScrollPositions);
      
      // Analyze the results
      const stickyElements = afterScrollPositions.filter(pos => pos.visible && pos.y !== null && pos.y < 100);
      console.log('Elements that appear sticky (near top):', stickyElements.length);
      
      const hiddenElements = afterScrollPositions.filter(pos => !pos.visible);
      console.log('Elements that became hidden:', hiddenElements.length);
      
      if (hiddenElements.length === afterScrollPositions.length) {
        console.log('❌ ALL sidebar elements are hidden after scrolling - this suggests they were pushed off screen');
      } else if (stickyElements.length > 0) {
        console.log('✅ Some sidebar elements remain visible and sticky');
      }
    } else {
      console.log('❌ No sidebar elements found at all');
    }
  });
});