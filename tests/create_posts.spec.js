const { test, expect } = require('@playwright/test');
const fs = require('fs');

test('Create three blog posts', async ({ page }) => {
  // Login
  await page.goto('http://localhost:22222/signin');
  await page.fill('input[name="email"]', 'anchoo2kewl@gmail.com');
  await page.fill('input[name="password"]', '123456qwertyu');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('networkidle');
  
  console.log('Logged in successfully');
  
  const posts = [
    {
      title: "Optimizing Cloud Resource Allocation with Machine Learning",
      slug: "optimizing-cloud-resource-allocation-ml",
      contentFile: "/tmp/post1_content.txt",
      category: "2" // Cloud Computing
    },
    {
      title: "Building Resilient Distributed Systems",
      slug: "building-resilient-distributed-systems", 
      contentFile: "/tmp/post2_content.txt",
      category: "3" // Architecture
    },
    {
      title: "The Future of Cloud Middleware Performance",
      slug: "future-cloud-middleware-performance",
      contentFile: "/tmp/post3_content.txt", 
      category: "4" // Performance
    }
  ];
  
  for (let i = 0; i < posts.length; i++) {
    const post = posts[i];
    console.log(`Creating post ${i + 1}: ${post.title}`);
    
    // Navigate to new post page
    await page.goto('http://localhost:22222/admin/posts/new');
    await page.waitForLoadState('networkidle');
    
    // Fill basic details
    await page.fill('input[name="title"]', post.title);
    await page.fill('input[name="slug"]', post.slug);
    
    // Read and set content
    const content = fs.readFileSync(post.contentFile, 'utf-8');
    await page.waitForSelector('#editor', { state: 'visible' });
    
    // Clear and set content
    await page.locator('#editor').click();
    await page.keyboard.press('Control+a');
    await page.locator('#editor').fill(content);
    
    // Set featured image
    await page.fill('input[name="featured_image_url"]', '/static/placeholder-featured.svg');
    
    // Select category if checkbox exists
    const categoryCheckbox = page.locator(`input[name="categories"][value="${post.category}"]`);
    if (await categoryCheckbox.isVisible().catch(() => false)) {
      await categoryCheckbox.check();
    }
    
    // Publish the post
    await page.check('input[name="is_published"]');
    
    // Submit
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');
    
    console.log(`✅ Created: ${post.title}`);
    
    // Wait a moment between posts
    await page.waitForTimeout(1000);
  }
  
  console.log('All posts created successfully!');
});