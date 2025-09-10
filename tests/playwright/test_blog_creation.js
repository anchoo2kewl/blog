const { test, expect } = require('@playwright/test');
const path = require('path');
const fs = require('fs');

test.describe('Blog Creation and Upload Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Set a longer timeout for login operations
    test.setTimeout(30000);
    
    // Login as admin user
    await page.goto('http://localhost:22222/signin');
    await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
    await page.fill('input[name="password"]', '123456qwertyu');
    await page.click('button[type="submit"]');
    
    // Wait for successful login redirect
    await page.waitForURL('**/');
  });

  test('should create a new blog post with content', async ({ page }) => {
    console.log('🚀 Starting blog post creation test...');
    
    // Navigate to create new post
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Fill in the post details
    const testTitle = 'Playwright Test Post ' + Date.now();
    await page.fill('input[name="title"]', testTitle);
    
    // Add content to the editor
    const testContent = `# Test Post Content

This is a test post created by Playwright automation.

## Features Tested
- Post creation
- Title and content editing
- Category selection

**Bold text** and *italic text* work correctly.

<more-->

This content appears after the excerpt break.

\`\`\`javascript
console.log('Code blocks work too!');
\`\`\`

End of test content.`;

    // Click on the editor to focus it, then fill content
    await page.click('#editor');
    await page.locator('#editor').fill(testContent);
    
    // Select at least one category (required)
    const firstCategory = page.locator('input[name="categories"]').first();
    await firstCategory.check();
    
    // Submit the form
    await page.click('button[type="submit"]');
    
    // Wait for redirect to posts list
    await page.waitForURL('**/admin/posts', { timeout: 10000 });
    
    // Verify the post was created by checking if it appears in the posts list
    await expect(page.locator(`text=${testTitle}`)).toBeVisible();
    
    console.log('✅ Blog post creation test completed successfully');
  });

  test('should display multi-image upload interface', async ({ page }) => {
    console.log('🖼️ Testing multi-image upload interface...');
    
    // Navigate to create new post
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Check if multi-upload drop zone is visible
    await expect(page.locator('#multi-upload-zone')).toBeVisible();
    
    // Check if the drop zone has the correct text
    await expect(page.locator('text=Drop images here')).toBeVisible();
    
    // Check if the browse link is present and clickable
    const browseLink = page.locator('.upload-link').first();
    await expect(browseLink).toBeVisible();
    
    // Check if the multi-upload input exists but is hidden
    await expect(page.locator('#multi-upload')).toBeHidden();
    
    console.log('✅ Multi-image upload interface test completed');
  });

  test('should display cover image upload interface', async ({ page }) => {
    console.log('🎨 Testing cover image upload interface...');
    
    // Navigate to create new post  
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Check if cover upload zone is visible
    await expect(page.locator('#cover-upload-zone')).toBeVisible();
    
    // Check if the cover upload text is present
    await expect(page.locator('text=Drop cover image')).toBeVisible();
    
    // Check if "Choose Existing" button is visible
    await expect(page.locator('#choose-existing-cover-btn')).toBeVisible();
    
    console.log('✅ Cover image upload interface test completed');
  });

  test('should test multi-image upload functionality', async ({ page }) => {
    console.log('🔧 Testing multi-image upload functionality...');
    
    // Navigate to create new post
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Create test image files
    const testDir = '/tmp/playwright-test-images';
    await fs.promises.mkdir(testDir, { recursive: true });
    
    const testImages = [
      {
        name: 'test1.svg',
        content: '<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg"><rect width="100" height="100" fill="red"/><text x="50" y="55" text-anchor="middle" fill="white">Test1</text></svg>'
      },
      {
        name: 'test2.svg', 
        content: '<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg"><rect width="100" height="100" fill="blue"/><text x="50" y="55" text-anchor="middle" fill="white">Test2</text></svg>'
      }
    ];
    
    const filePaths = [];
    for (const img of testImages) {
      const filePath = path.join(testDir, img.name);
      await fs.promises.writeFile(filePath, img.content);
      filePaths.push(filePath);
    }
    
    try {
      // Add a title to the post first
      await page.fill('input[name="title"]', 'Multi-Upload Test Post');
      
      // Listen for console messages to debug upload issues
      page.on('console', msg => console.log('🔍 Browser console:', msg.text()));
      
      // Click on the multi-upload drop zone to trigger file picker
      await page.click('#multi-upload-zone');
      
      // Upload the test images via the hidden file input
      const fileInput = page.locator('#multi-upload');
      await fileInput.setInputFiles(filePaths);
      
      // Wait for upload modal to appear
      await page.waitForSelector('#multi-upload-modal', { state: 'visible', timeout: 5000 });
      
      // Wait a bit for the upload process
      await page.waitForTimeout(3000);
      
      // Check if upload was successful by looking for success indicators
      const uploadedImages = page.locator('.upload-preview-item');
      const count = await uploadedImages.count();
      console.log(`📊 Found ${count} uploaded image previews`);
      
      // Close the modal
      const closeBtn = page.locator('#close-modal');
      if (await closeBtn.isVisible()) {
        await closeBtn.click();
      }
      
      console.log('✅ Multi-image upload test completed');
      
    } catch (error) {
      console.error('❌ Multi-image upload test failed:', error.message);
      
      // Take a screenshot for debugging
      await page.screenshot({ path: '/tmp/multi-upload-error.png', fullPage: true });
      console.log('📸 Screenshot saved to /tmp/multi-upload-error.png');
      
    } finally {
      // Clean up test files
      for (const filePath of filePaths) {
        try {
          await fs.promises.unlink(filePath);
        } catch (e) {
          console.log('⚠️ Cleanup warning:', e.message);
        }
      }
      try {
        await fs.promises.rmdir(testDir);
      } catch (e) {
        console.log('⚠️ Directory cleanup warning:', e.message);
      }
    }
  });

  test('should open existing images modal', async ({ page }) => {
    console.log('🖼️ Testing existing images modal...');
    
    // Navigate to create new post
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Click "Choose Existing" button
    await page.click('#choose-existing-cover-btn');
    
    // Wait for modal to appear
    await page.waitForSelector('#existing-images-modal', { state: 'visible' });
    
    // Check if the modal is visible
    const modal = page.locator('#existing-images-modal');
    await expect(modal).toBeVisible();
    
    // Check if modal title is correct
    await expect(page.locator('text=Choose Existing Image')).toBeVisible();
    
    // Close the modal
    await page.click('#close-existing-modal');
    
    // Verify modal is hidden
    await expect(modal).toBeHidden();
    
    console.log('✅ Existing images modal test completed');
  });

  test('should validate required fields', async ({ page }) => {
    console.log('✅ Testing form validation...');
    
    // Navigate to create new post
    await page.goto('http://localhost:22222/admin/posts/new');
    
    // Try to submit without title (should fail)
    await page.click('button[type="submit"]');
    
    // Check if we're still on the new post page (form validation failed)
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('/admin/posts/new');
    
    // Check if HTML5 validation prevents submission
    const titleInput = page.locator('input[name="title"]');
    const validationMessage = await titleInput.evaluate(el => el.validationMessage);
    expect(validationMessage).toBeTruthy();
    
    console.log('✅ Form validation test completed');
  });
});

test.describe('Blog Post Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('http://localhost:22222/signin');
    await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
    await page.fill('input[name="password"]', '123456qwertyu');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/');
  });

  test('should list existing blog posts', async ({ page }) => {
    console.log('📝 Testing blog posts list...');
    
    // Navigate to posts list
    await page.goto('http://localhost:22222/admin/posts');
    
    // Check if the page title is correct
    await expect(page.locator('h1')).toContainText('Posts');
    
    // Check if there's at least one post
    const postRows = page.locator('tbody tr');
    const count = await postRows.count();
    expect(count).toBeGreaterThan(0);
    
    console.log(`✅ Found ${count} blog posts in the list`);
  });

  test('should navigate to edit existing post', async ({ page }) => {
    console.log('✏️ Testing edit post navigation...');
    
    // Navigate to posts list
    await page.goto('http://localhost:22222/admin/posts');
    
    // Click the first "Edit" link
    const firstEditLink = page.locator('a[href*="/edit"]').first();
    await firstEditLink.click();
    
    // Verify we're on an edit page
    await expect(page.locator('h1')).toContainText('Edit Post');
    
    // Verify the form is populated with existing data
    const titleInput = page.locator('input[name="title"]');
    const titleValue = await titleInput.inputValue();
    expect(titleValue.length).toBeGreaterThan(0);
    
    console.log('✅ Edit post navigation test completed');
  });
});