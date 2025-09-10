const { test, expect } = require('@playwright/test');
const path = require('path');

test.describe('Editor Layout Tests (No Auth)', () => {
  const testFile = 'file://' + path.resolve(__dirname, 'layout-test.html');

  test('should display proper layout structure on desktop', async ({ page }) => {
    // Set viewport to desktop size
    await page.setViewportSize({ width: 1200, height: 800 });
    
    // Navigate to test file
    await page.goto(testFile);
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Check that layout container exists
    const editorLayout = page.locator('.editor-layout');
    await expect(editorLayout).toBeVisible();
    
    // Check that main section exists and is visible
    const editorMain = page.locator('.editor-main');
    await expect(editorMain).toBeVisible();
    
    // Check that sidebar exists and is visible
    const editorSidebar = page.locator('.editor-sidebar');
    await expect(editorSidebar).toBeVisible();
    
    console.log('✅ All layout elements are visible');
  });

  test('should maintain proper desktop layout proportions', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    const editorMain = page.locator('.editor-main');
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get bounding boxes
    const mainBox = await editorMain.boundingBox();
    const sidebarBox = await editorSidebar.boundingBox();
    
    console.log('Main section:', mainBox);
    console.log('Sidebar section:', sidebarBox);
    
    // Verify layout is horizontal (side by side)
    expect(mainBox.x).toBeLessThan(sidebarBox.x);
    console.log('✅ Sidebar is to the right of main content');
    
    // Verify sidebar has fixed width around 320px (allowing some flexibility)
    expect(sidebarBox.width).toBeGreaterThan(300);
    expect(sidebarBox.width).toBeLessThan(380);
    console.log(`✅ Sidebar width is ${sidebarBox.width}px (expected ~320-370px)`);
    
    // Verify main section is wider than sidebar
    expect(mainBox.width).toBeGreaterThan(sidebarBox.width);
    const ratio = mainBox.width / sidebarBox.width;
    expect(ratio).toBeGreaterThan(1.5); // Should be roughly 2:1
    console.log(`✅ Width ratio is ${ratio.toFixed(2)}:1 (main:sidebar)`);
  });

  test('should make sidebar sticky when scrolling', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get initial position
    const initialBox = await editorSidebar.boundingBox();
    console.log('Initial sidebar position:', initialBox.y);
    
    // Scroll down significantly
    await page.evaluate(() => {
      window.scrollTo(0, 1000);
    });
    
    // Wait for scroll to complete
    await page.waitForTimeout(500);
    
    // Get position after scroll
    const scrolledBox = await editorSidebar.boundingBox();
    console.log('Sidebar position after scroll:', scrolledBox.y);
    
    // Sidebar should have moved up relative to its initial position
    // (indicating it's sticky and following the scroll)
    expect(scrolledBox.y).toBeLessThan(initialBox.y);
    console.log('✅ Sidebar moved up when scrolling (sticky behavior)');
    
    // But it should still be visible on screen (not scrolled off)
    expect(scrolledBox.y).toBeGreaterThan(-50);
    console.log('✅ Sidebar remains visible on screen');
  });

  test('should stack vertically on mobile', async ({ page }) => {
    // Set viewport to mobile size
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    const editorMain = page.locator('.editor-main');
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get bounding boxes
    const mainBox = await editorMain.boundingBox();
    const sidebarBox = await editorSidebar.boundingBox();
    
    console.log('Mobile - Main section:', mainBox);
    console.log('Mobile - Sidebar section:', sidebarBox);
    
    // On mobile, sidebar should be below main content
    expect(sidebarBox.y).toBeGreaterThan(mainBox.y + mainBox.height - 100); // Allow some overlap/margin
    console.log('✅ Sidebar is below main content on mobile');
    
    // Both should take most of the available width
    const viewportWidth = 375;
    expect(mainBox.width).toBeGreaterThan(viewportWidth * 0.7); // Allow for padding
    expect(sidebarBox.width).toBeGreaterThan(viewportWidth * 0.7);
    console.log(`✅ Both sections use full width (main: ${mainBox.width}px, sidebar: ${sidebarBox.width}px)`);
  });

  test('should have all form elements accessible', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    const sidebar = page.locator('.editor-sidebar');
    
    // Check slug input
    const slugInput = sidebar.locator('input[placeholder*="Slug"]');
    await expect(slugInput).toBeVisible();
    await expect(slugInput).toHaveValue('the-ultimate-guide-to-modern-web-development-a-comprehensive-journey');
    console.log('✅ Slug input is visible and has value');
    
    // Check featured image input
    const imageInput = sidebar.locator('input[placeholder*="Featured Image"]');
    await expect(imageInput).toBeVisible();
    await expect(imageInput).toHaveValue('/static/placeholder-featured.svg');
    console.log('✅ Featured image input is visible and has value');
    
    // Check categories checkboxes
    const categoryCheckboxes = sidebar.locator('input[type="checkbox"]');
    const count = await categoryCheckboxes.count();
    expect(count).toBeGreaterThan(3); // Should have multiple category options
    console.log(`✅ Found ${count} category checkboxes`);
    
    // Check submit button
    const submitButton = sidebar.locator('button[type="submit"]');
    await expect(submitButton).toBeVisible();
    await expect(submitButton).toContainText('Update Post');
    console.log('✅ Submit button is visible');
  });

  test('should handle content editor interactions', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    const contentEditor = page.locator('.wysiwyg');
    await expect(contentEditor).toBeVisible();
    
    // Check that content is loaded
    const content = await contentEditor.textContent();
    expect(content.length).toBeGreaterThan(1000);
    console.log(`✅ Content editor has ${content.length} characters`);
    
    // Check that editor is editable
    await contentEditor.click();
    await page.keyboard.press('End'); // Go to end of content
    await page.keyboard.type('\n\nThis is a test edit!');
    
    // Verify the text was added
    const updatedContent = await contentEditor.textContent();
    expect(updatedContent).toContain('This is a test edit!');
    console.log('✅ Content editor is editable');
  });

  test('should have responsive breakpoints working', async ({ page }) => {
    await page.goto(testFile);
    await page.waitForLoadState('networkidle');
    
    // Test different viewport sizes
    const viewports = [
      { width: 1200, height: 800, name: 'Desktop' },
      { width: 768, height: 1024, name: 'Tablet' },
      { width: 375, height: 667, name: 'Mobile' }
    ];
    
    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(300); // Allow layout to adjust
      
      const editorMain = page.locator('.editor-main');
      const editorSidebar = page.locator('.editor-sidebar');
      
      const mainBox = await editorMain.boundingBox();
      const sidebarBox = await editorSidebar.boundingBox();
      
      const isSideBySide = mainBox.x < sidebarBox.x;
      const isStacked = sidebarBox.y > mainBox.y + mainBox.height - 100;
      
      console.log(`${viewport.name} (${viewport.width}x${viewport.height}):`);
      console.log(`  Side by side: ${isSideBySide}`);
      console.log(`  Stacked: ${isStacked}`);
      console.log(`  Main: ${mainBox.width}x${mainBox.height} at (${mainBox.x}, ${mainBox.y})`);
      console.log(`  Sidebar: ${sidebarBox.width}x${sidebarBox.height} at (${sidebarBox.x}, ${sidebarBox.y})`);
      
      if (viewport.width >= 768) {
        expect(isSideBySide).toBe(true);
        console.log(`  ✅ ${viewport.name} uses side-by-side layout`);
      } else {
        expect(isStacked).toBe(true);
        console.log(`  ✅ ${viewport.name} uses stacked layout`);
      }
    }
  });
});