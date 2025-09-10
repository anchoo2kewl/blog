const { test, expect } = require('@playwright/test');

test('Check post count after server restart', async ({ page }) => {
  await page.goto('http://localhost:22222');
  await page.waitForLoadState('networkidle');
  
  // Count articles (posts)
  const articles = page.locator('article');
  const articleCount = await articles.count();
  
  console.log(`Number of articles displayed: ${articleCount}`);
  
  // Get all post titles
  for (let i = 0; i < articleCount; i++) {
    const article = articles.nth(i);
    const titleElement = article.locator('h3, h2, .title');
    const title = await titleElement.textContent();
    console.log(`Post ${i + 1}: ${title?.trim()}`);
  }
  
  // Check if our 3 technical posts are visible
  const technicalPosts = [
    'Optimizing Cloud Resource Allocation with Machine Learning',
    'Building Resilient Distributed Systems', 
    'The Future of Cloud Middleware Performance'
  ];
  
  const pageText = await page.textContent('body');
  
  for (const postTitle of technicalPosts) {
    const isVisible = pageText.includes(postTitle);
    console.log(`"${postTitle}" visible: ${isVisible}`);
  }
  
  // We should now see more than 5 posts
  expect(articleCount).toBeGreaterThanOrEqual(5);
});