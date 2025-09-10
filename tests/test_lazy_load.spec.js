const { test, expect } = require('@playwright/test');

test.describe('Homepage Post Display and Lazy Loading', () => {
  test('should check current number of posts displayed', async ({ page }) => {
    await page.goto('http://localhost:22222');
    await page.waitForLoadState('networkidle');
    
    // Count the number of post cards on the homepage
    const postCards = page.locator('[class*="post"], article, .post-card, .post-item');
    const postCount = await postCards.count();
    
    console.log(`Number of post elements found: ${postCount}`);
    
    // Try different selectors to find posts
    const alternativeSelectors = [
      'h3 a[href*="/posts/"]',
      'h2 a[href*="/posts/"]', 
      'a[href*="/posts/"]:has(h2,h3)',
      '[href*="/posts/"]',
      '.post',
      'article'
    ];
    
    for (const selector of alternativeSelectors) {
      const elements = page.locator(selector);
      const count = await elements.count();
      if (count > 0) {
        console.log(`Selector "${selector}" found ${count} elements`);
        
        // Get the first few post titles to verify
        for (let i = 0; i < Math.min(count, 5); i++) {
          const element = elements.nth(i);
          const text = await element.textContent();
          const href = await element.getAttribute('href');
          console.log(`Post ${i + 1}: "${text?.trim()}" -> ${href}`);
        }
      }
    }
    
    // Check if there are any "Load More" or pagination elements
    const loadMoreButton = page.locator('button:has-text("Load More"), button:has-text("Show More"), .load-more, .pagination');
    const hasLoadMore = await loadMoreButton.isVisible().catch(() => false);
    
    console.log(`Load more/pagination visible: ${hasLoadMore}`);
    
    // Check the page source for post data
    const pageContent = await page.content();
    const postMatches = pageContent.match(/href="[^"]*\/posts\/[^"]*"/g) || [];
    console.log(`Post links found in HTML: ${postMatches.length}`);
    console.log('Post links:', postMatches.slice(0, 10));
  });
  
  test('should test lazy loading if implemented', async ({ page }) => {
    await page.goto('http://localhost:22222');
    await page.waitForLoadState('networkidle');
    
    // Check for initial posts
    const initialPosts = page.locator('a[href*="/posts/"]');
    const initialCount = await initialPosts.count();
    console.log(`Initial posts loaded: ${initialCount}`);
    
    // Look for load more functionality
    const loadMoreSelectors = [
      'button:has-text("Load More")',
      'button:has-text("Show More")', 
      '.load-more',
      'button[onclick*="load"]',
      'button[id*="load"]',
      '[data-action="load-more"]'
    ];
    
    let loadMoreButton = null;
    for (const selector of loadMoreSelectors) {
      const button = page.locator(selector);
      if (await button.isVisible().catch(() => false)) {
        loadMoreButton = button;
        console.log(`Found load more button with selector: ${selector}`);
        break;
      }
    }
    
    if (loadMoreButton) {
      // Test clicking load more
      console.log('Testing load more functionality...');
      await loadMoreButton.click();
      await page.waitForTimeout(2000); // Wait for content to load
      
      const newCount = await initialPosts.count();
      console.log(`Posts after load more: ${newCount}`);
      
      if (newCount > initialCount) {
        console.log('✅ Lazy loading is working!');
      } else {
        console.log('❌ Lazy loading did not increase post count');
      }
    } else {
      console.log('No load more button found - checking if infinite scroll exists');
      
      // Test infinite scroll by scrolling to bottom
      const initialPostCount = await initialPosts.count();
      
      // Scroll to bottom
      await page.evaluate(() => {
        window.scrollTo(0, document.body.scrollHeight);
      });
      
      await page.waitForTimeout(2000);
      
      const afterScrollCount = await initialPosts.count();
      if (afterScrollCount > initialPostCount) {
        console.log('✅ Infinite scroll is working!');
      } else {
        console.log('❌ No lazy loading detected (neither button nor infinite scroll)');
      }
    }
  });
  
  test('should analyze page structure for debugging', async ({ page }) => {
    await page.goto('http://localhost:22222');
    await page.waitForLoadState('networkidle');
    
    // Get the main content structure
    const bodyHTML = await page.locator('body').innerHTML();
    
    // Look for post-related patterns
    const patterns = [
      /class="[^"]*post[^"]*"/gi,
      /href="[^"]*\/posts\/[^"]*"/gi,
      /\/admin\/posts\/\d+/gi,
      /post-\d+/gi
    ];
    
    console.log('=== PAGE STRUCTURE ANALYSIS ===');
    
    for (const pattern of patterns) {
      const matches = bodyHTML.match(pattern) || [];
      if (matches.length > 0) {
        console.log(`Pattern ${pattern}: ${matches.length} matches`);
        console.log('Examples:', matches.slice(0, 3));
      }
    }
    
    // Check if we can find any containers that might hold posts
    const containerSelectors = ['main', '.container', '.content', '.posts', '.blog-posts', 'section'];
    
    for (const selector of containerSelectors) {
      const container = page.locator(selector).first();
      if (await container.isVisible().catch(() => false)) {
        const containerHTML = await container.innerHTML();
        const postLinks = containerHTML.match(/href="[^"]*\/posts\/[^"]*"/g) || [];
        if (postLinks.length > 0) {
          console.log(`Container "${selector}" contains ${postLinks.length} post links`);
        }
      }
    }
  });
});