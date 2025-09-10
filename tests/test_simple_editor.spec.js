const { test, expect } = require('@playwright/test');

test('Test editor layout with a shorter post', async ({ page }) => {
  // Login first
  await page.goto('/signin');
  await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
  await page.fill('input[name="password"]', '123456qwertyu');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('networkidle');
  
  // Navigate to a different post (post 9 which has less content)
  await page.goto('/admin/posts/9/edit');
  await page.waitForLoadState('networkidle');
  
  // Set desktop viewport
  await page.setViewportSize({ width: 1200, height: 800 });
  
  // Wait for layout to settle
  await page.waitForTimeout(2000);
  
  console.log('Current URL:', page.url());
  
  // Check for console messages (JavaScript execution)
  const consoleLogs = [];
  page.on('console', msg => consoleLogs.push(msg.text()));
  
  // Wait a bit more to catch any console messages
  await page.waitForTimeout(1000);
  
  console.log('Console messages:', consoleLogs.filter(log => log.includes('Layout')));
  
  // Check the actual layout
  const layoutElement = page.locator('.editor-layout').first();
  const mainElement = page.locator('.editor-main').first();
  const sidebarElement = page.locator('.editor-sidebar').first();
  
  if (await layoutElement.isVisible()) {
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
    
    if (await mainElement.isVisible()) {
      const mainStyles = await mainElement.evaluate(el => {
        const styles = getComputedStyle(el);
        return {
          display: styles.display,
          flex: styles.flex,
          position: styles.position,
          width: styles.width,
          order: styles.order
        };
      });
      console.log('Main element styles:', mainStyles);
    }
    
    if (await sidebarElement.isVisible()) {
      const sidebarStyles = await sidebarElement.evaluate(el => {
        const styles = getComputedStyle(el);
        return {
          display: styles.display,
          flex: styles.flex,
          position: styles.position,
          width: styles.width,
          order: styles.order
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
        const isSideBySide = Math.abs(sidebarBox.y - mainBox.y) < 100; // Same row
        const sidebarFirst = sidebarBox.x < mainBox.x; // Sidebar on left
        
        console.log('Layout analysis:');
        console.log('- Side by side:', isSideBySide);
        console.log('- Sidebar first (left):', sidebarFirst);
        console.log('- Stacked (bad):', isStacked);
        
        if (isStacked) {
          console.log('❌ LAYOUT ISSUE: Sidebar pushed to bottom');
        } else if (isSideBySide && sidebarFirst) {
          console.log('✅ Layout correct - sidebar on left, main on right');
        } else if (isSideBySide && !sidebarFirst) {
          console.log('⚠️ Layout partially correct but sidebar on right instead of left');
        } else {
          console.log('⚠️ Unexpected layout');
        }
      }
    }
  } else {
    console.log('❌ Layout elements not found');
  }
});