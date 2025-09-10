const { test, expect } = require('@playwright/test');

test.describe('Post Editor Layout Tests', () => {
  let page;

  test.beforeAll(async ({ browser }) => {
    page = await browser.newPage();
  });

  test.afterAll(async () => {
    await page.close();
  });

  test('should login as admin user', async () => {
    // Navigate to login page
    await page.goto('http://localhost:22222/signin');
    
    // Fill login form (adjust these credentials as needed)
    await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
    await page.fill('input[name="password"]', '123456qwertyu'); // Adjust password as needed
    
    // Submit login
    await page.click('button[type="submit"]');
    
    // Wait for redirect and verify we're logged in
    await page.waitForURL('**/admin**', { timeout: 5000 });
    
    // Verify we're logged in by checking for admin navigation
    const adminElement = page.locator('text=Admin');
    await expect(adminElement).toBeVisible();
  });

  test('should navigate to post editor', async () => {
    // Go to posts list
    await page.goto('http://localhost:22222/admin/posts');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Click on edit button for the comprehensive post (ID 8)
    const editButton = page.locator('a[href="/admin/posts/8/edit"]').first();
    await expect(editButton).toBeVisible();
    await editButton.click();
    
    // Wait for editor page to load
    await page.waitForLoadState('networkidle');
    await page.waitForSelector('.editor-layout', { timeout: 10000 });
    
    // Verify we're on the editor page
    await expect(page.locator('h1')).toContainText('Edit Post');
  });

  test('should have proper editor layout structure', async () => {
    // Check that editor layout container exists
    const editorLayout = page.locator('.editor-layout');
    await expect(editorLayout).toBeVisible();
    
    // Check that main editor section exists and is visible
    const editorMain = page.locator('.editor-main');
    await expect(editorMain).toBeVisible();
    
    // Check that sidebar exists and is visible
    const editorSidebar = page.locator('.editor-sidebar');
    await expect(editorSidebar).toBeVisible();
    
    // Verify title input is in main section
    const titleInput = editorMain.locator('input[name="title"]');
    await expect(titleInput).toBeVisible();
    
    // Verify content editor is in main section
    const contentEditor = editorMain.locator('#editor');
    await expect(contentEditor).toBeVisible();
  });

  test('should have metadata controls in sidebar', async () => {
    const sidebar = page.locator('.editor-sidebar');
    
    // Check slug input
    const slugInput = sidebar.locator('input[name="slug"]');
    await expect(slugInput).toBeVisible();
    
    // Check featured image URL input
    const featuredImageInput = sidebar.locator('input[name="featured_image_url"]');
    await expect(featuredImageInput).toBeVisible();
    
    // Check categories section
    const categoriesSection = sidebar.locator('text=Categories');
    await expect(categoriesSection).toBeVisible();
    
    // Check publish checkbox
    const publishCheckbox = sidebar.locator('input[name="is_published"]');
    await expect(publishCheckbox).toBeVisible();
    
    // Check submit button
    const submitButton = sidebar.locator('button[type="submit"]');
    await expect(submitButton).toBeVisible();
    await expect(submitButton).toContainText('Update Post');
  });

  test('should maintain layout proportions on desktop', async () => {
    // Set viewport to desktop size
    await page.setViewportSize({ width: 1200, height: 800 });
    
    const editorLayout = page.locator('.editor-layout');
    const editorMain = page.locator('.editor-main');
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get bounding boxes
    const layoutBox = await editorLayout.boundingBox();
    const mainBox = await editorMain.boundingBox();
    const sidebarBox = await editorSidebar.boundingBox();
    
    // Verify layout is horizontal (side by side)
    expect(mainBox.x).toBeLessThan(sidebarBox.x);
    
    // Verify main section takes more space than sidebar (roughly 2:1 ratio)
    const mainWidth = mainBox.width;
    const sidebarWidth = sidebarBox.width;
    const ratio = mainWidth / sidebarWidth;
    
    // Should be roughly 2:1 ratio (allow some tolerance)
    expect(ratio).toBeGreaterThan(1.5);
    expect(ratio).toBeLessThan(3.0);
    
    // Verify sidebar has fixed width around 320px
    expect(sidebarWidth).toBeGreaterThan(300);
    expect(sidebarWidth).toBeLessThan(350);
  });

  test('should make sidebar sticky when scrolling', async () => {
    // Set viewport to desktop size
    await page.setViewportSize({ width: 1200, height: 800 });
    
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get initial position
    const initialBox = await editorSidebar.boundingBox();
    const initialTop = initialBox.y;
    
    // Scroll down significantly
    await page.evaluate(() => {
      window.scrollTo(0, 1000);
    });
    
    // Wait for scroll to complete
    await page.waitForTimeout(500);
    
    // Get position after scroll
    const scrolledBox = await editorSidebar.boundingBox();
    const scrolledTop = scrolledBox.y;
    
    // Sidebar should have moved up relative to its initial position
    // (indicating it's sticky and following the scroll)
    expect(scrolledTop).toBeLessThan(initialTop);
    
    // But it should still be visible on screen
    expect(scrolledTop).toBeGreaterThan(-100); // Allow some margin
  });

  test('should stack vertically on mobile', async () => {
    // Set viewport to mobile size
    await page.setViewportSize({ width: 375, height: 667 });
    
    const editorMain = page.locator('.editor-main');
    const editorSidebar = page.locator('.editor-sidebar');
    
    // Get bounding boxes
    const mainBox = await editorMain.boundingBox();
    const sidebarBox = await editorSidebar.boundingBox();
    
    // On mobile, sidebar should be below main content
    expect(sidebarBox.y).toBeGreaterThan(mainBox.y + mainBox.height - 50); // Allow some overlap
    
    // Both should take full width
    const viewportWidth = 375;
    expect(mainBox.width).toBeGreaterThan(viewportWidth * 0.8); // Allow for padding
    expect(sidebarBox.width).toBeGreaterThan(viewportWidth * 0.8);
  });

  test('should handle long content without breaking layout', async () => {
    // Set viewport to desktop size
    await page.setViewportSize({ width: 1200, height: 800 });
    
    // Get current content length
    const contentEditor = page.locator('#editor');
    const initialContent = await contentEditor.textContent();
    
    console.log(`Initial content length: ${initialContent.length}`);
    
    // Verify the editor contains the long comprehensive content
    expect(initialContent.length).toBeGreaterThan(10000); // Should be a very long post
    
    // Check that layout is still intact with long content
    const editorLayout = page.locator('.editor-layout');
    const editorMain = page.locator('.editor-main');
    const editorSidebar = page.locator('.editor-sidebar');
    
    // All elements should still be visible
    await expect(editorLayout).toBeVisible();
    await expect(editorMain).toBeVisible();
    await expect(editorSidebar).toBeVisible();
    
    // Get positions to verify layout
    const mainBox = await editorMain.boundingBox();
    const sidebarBox = await editorSidebar.boundingBox();
    
    // Sidebar should still be positioned to the right of main content
    expect(sidebarBox.x).toBeGreaterThan(mainBox.x + mainBox.width - 50); // Allow some margin
    
    // Sidebar should be at top of viewport area (sticky)
    expect(sidebarBox.y).toBeLessThan(100);
  });

  test('should preserve content when switching between edit and preview', async () => {
    // Click preview tab
    const previewTab = page.locator('#tab-preview');
    await previewTab.click();
    
    // Wait for preview to load
    await page.waitForSelector('#preview:not(.hidden)', { timeout: 5000 });
    
    // Verify preview is visible and edit is hidden
    const preview = page.locator('#preview');
    const editor = page.locator('#editor');
    
    await expect(preview).toBeVisible();
    await expect(editor).toBeHidden();
    
    // Click edit tab to go back
    const editTab = page.locator('#tab-edit');
    await editTab.click();
    
    // Verify editor is visible again
    await expect(editor).toBeVisible();
    await expect(preview).toBeHidden();
    
    // Verify content is still there
    const content = await editor.textContent();
    expect(content.length).toBeGreaterThan(1000);
  });

  test('should show proper visual feedback for interactive elements', async () => {
    // Test toolbar buttons are clickable
    const boldButton = page.locator('button[data-cmd="bold"]');
    await expect(boldButton).toBeVisible();
    
    // Test sidebar controls
    const slugInput = page.locator('input[name="slug"]');
    await expect(slugInput).toBeEnabled();
    
    // Test submit button styling
    const submitButton = page.locator('button[type="submit"]');
    await expect(submitButton).toBeVisible();
    await expect(submitButton).toBeEnabled();
    
    // Test that sidebar has proper styling (background color, etc.)
    const sidebar = page.locator('.editor-sidebar');
    const bgColor = await sidebar.evaluate(el => getComputedStyle(el).backgroundColor);
    expect(bgColor).not.toBe('rgba(0, 0, 0, 0)'); // Should have a background color
  });
});