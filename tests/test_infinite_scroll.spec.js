const { test, expect } = require('@playwright/test');

test.describe('Infinite Scroll Functionality', () => {
  test('should initially show 5 posts and load more on scroll', async ({ page }) => {
    await page.goto('http://localhost:22222');
    await page.waitForLoadState('networkidle');
    
    // Count initial posts
    const initialArticles = page.locator('article.blog-post');
    const initialCount = await initialArticles.count();
    console.log(`Initial posts loaded: ${initialCount}`);
    
    // Should start with 5 posts
    expect(initialCount).toBe(5);
    
    // Get titles of first 5 posts for verification
    const initialTitles = [];
    for (let i = 0; i < initialCount; i++) {
      const title = await initialArticles.nth(i).locator('h2.blog-post-title a').textContent();
      initialTitles.push(title.trim());
    }
    console.log('Initial posts:', initialTitles);
    
    // Scroll down to trigger infinite scroll
    await page.evaluate(() => {
      window.scrollTo(0, document.body.scrollHeight);
    });
    
    // Wait for new posts to load
    await page.waitForTimeout(2000);
    
    // Count posts after scroll
    const afterScrollArticles = page.locator('article.blog-post');
    const afterScrollCount = await afterScrollArticles.count();
    console.log(`Posts after scroll: ${afterScrollCount}`);
    
    // Should have more than 5 posts now
    expect(afterScrollCount).toBeGreaterThan(5);
    expect(afterScrollCount).toBeLessThanOrEqual(10); // Maximum 10 posts total
    
    // Verify that new posts were added (check last few posts)
    const allTitles = [];
    for (let i = 0; i < afterScrollCount; i++) {
      const title = await afterScrollArticles.nth(i).locator('h2.blog-post-title a').textContent();
      allTitles.push(title.trim());
    }
    console.log('All posts after scroll:', allTitles);
    
    // Verify our 3 technical posts from 2024 are now visible
    const technicalPosts = [
      'Optimizing Cloud Resource Allocation with Machine Learning',
      'Building Resilient Distributed Systems', 
      'The Future of Cloud Middleware Performance'
    ];
    
    for (const postTitle of technicalPosts) {
      const isVisible = allTitles.includes(postTitle);
      console.log(`"${postTitle}" visible: ${isVisible}`);
      expect(isVisible).toBe(true);
    }
  });
  
  test('should test API endpoint directly', async ({ page }) => {
    // Test the API endpoint directly
    const response = await page.request.get('http://localhost:22222/api/posts/load-more?offset=5');
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    console.log(`API returned ${data.Posts ? data.Posts.length : 0} posts`);
    
    if (data.Posts && data.Posts.length > 0) {
      console.log('First API post:', data.Posts[0].Title);
      console.log('API posts:', data.Posts.map(p => p.Title));
      
      // Should return the remaining posts (our 3 technical posts from 2024)
      expect(data.Posts.length).toBe(5);
      
      // Verify structure of returned posts
      const firstPost = data.Posts[0];
      expect(firstPost).toHaveProperty('ID');
      expect(firstPost).toHaveProperty('Title');
      expect(firstPost).toHaveProperty('Slug');
      expect(firstPost).toHaveProperty('Content');
      expect(firstPost).toHaveProperty('CreatedAt');
      expect(firstPost).toHaveProperty('PublicationDate');
    }
  });
  
  test('should handle multiple scroll events', async ({ page }) => {
    await page.goto('http://localhost:22222');
    await page.waitForLoadState('networkidle');
    
    // Start with 5 posts
    let articleCount = await page.locator('article.blog-post').count();
    console.log(`Start: ${articleCount} posts`);
    expect(articleCount).toBe(5);
    
    // First scroll - should load next 5 posts
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(2000);
    
    articleCount = await page.locator('article.blog-post').count();
    console.log(`After first scroll: ${articleCount} posts`);
    expect(articleCount).toBe(10);
    
    // Second scroll - should not load more (no more posts available)
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(2000);
    
    const finalCount = await page.locator('article.blog-post').count();
    console.log(`After second scroll: ${finalCount} posts`);
    expect(finalCount).toBe(10); // Should remain at 10
  });
});