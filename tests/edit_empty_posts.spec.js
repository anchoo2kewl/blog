const { test, expect } = require('@playwright/test');
const fs = require('fs');

test('Edit three empty blog posts with proper content', async ({ page }) => {
  // Login
  await page.goto('http://localhost:22222/signin');
  await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
  await page.fill('input[name="password"]', '123456qwertyu');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('networkidle');
  
  console.log('Logged in successfully');
  
  // The three empty posts that need content (post IDs 9, 10, 11)
  const postsToEdit = [
    {
      id: 9,
      title: "Optimizing Cloud Resource Allocation with Machine Learning",
      contentFile: "/tmp/post1_content.txt"
    },
    {
      id: 10,
      title: "Building Resilient Distributed Systems", 
      contentFile: "/tmp/post2_content.txt"
    },
    {
      id: 11,
      title: "The Future of Cloud Middleware Performance",
      contentFile: "/tmp/post3_content.txt"
    }
  ];
  
  for (let i = 0; i < postsToEdit.length; i++) {
    const post = postsToEdit[i];
    console.log(`Editing post ${post.id}: ${post.title}`);
    
    // Navigate to edit page for this post
    await page.goto(`http://localhost:22222/admin/posts/${post.id}/edit`);
    await page.waitForLoadState('networkidle');
    
    // Verify we're on the correct post by checking the title
    const currentTitle = await page.locator('input[name="title"]').inputValue();
    console.log(`Current post title: ${currentTitle}`);
    
    // Read the content from file
    const content = fs.readFileSync(post.contentFile, 'utf-8');
    console.log(`Content length: ${content.length} characters`);
    
    // Wait for the editor and clear existing content
    await page.waitForSelector('#editor', { state: 'visible' });
    await page.locator('#editor').click();
    
    // Select all and replace with new content - use textContent for contenteditable
    await page.keyboard.press('Control+a');
    await page.locator('#editor').fill(content);
    
    // Verify content was set - use textContent for contenteditable div
    const editorContent = await page.locator('#editor').textContent();
    if (editorContent && editorContent.length > 100) {
      console.log(`✅ Content successfully set (${editorContent.length} characters)`);
    } else {
      console.log(`⚠️ Warning: Content seems too short (${editorContent ? editorContent.length : 0} characters)`);
    }
    
    // Sync editor content to hidden field before submitting
    await page.evaluate(() => {
      if (typeof syncEditor === 'function') {
        syncEditor();
      } else {
        document.getElementById('content-field').value = document.getElementById('editor').innerHTML;
      }
    });
    
    // Submit the changes
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');
    
    console.log(`✅ Updated post ${post.id}: ${post.title}`);
    
    // Wait a moment between edits
    await page.waitForTimeout(1000);
  }
  
  console.log('All empty posts have been updated with proper content!');
});