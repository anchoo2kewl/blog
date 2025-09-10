const { test, expect } = require('@playwright/test');

test('Debug infinite scroll JavaScript', async ({ page }) => {
  // Enable console logging
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  page.on('pageerror', error => console.log('PAGE ERROR:', error.message));
  
  await page.goto('http://localhost:22222');
  await page.waitForLoadState('networkidle');
  
  // Check if the script variables are defined
  const jsCheck = await page.evaluate(() => {
    return {
      offsetExists: typeof window.offset !== 'undefined',
      loadingExists: typeof window.loading !== 'undefined',
      hasMorePostsExists: typeof window.hasMorePosts !== 'undefined',
      loadMorePostsExists: typeof window.loadMorePosts === 'function',
      offset: window.offset,
      loading: window.loading,
      hasMorePosts: window.hasMorePosts
    };
  });
  
  console.log('JavaScript variables:', jsCheck);
  
  // Check if scroll event listener is attached
  const scrollListenerCheck = await page.evaluate(() => {
    return {
      scrollHeight: document.body.scrollHeight,
      windowHeight: window.innerHeight,
      currentScroll: window.scrollY
    };
  });
  
  console.log('Scroll info:', scrollListenerCheck);
  
  // Manually trigger scroll and see what happens
  await page.evaluate(() => {
    console.log('Before manual scroll - offset:', window.offset, 'loading:', window.loading);
    
    // Manually call loadMorePosts to test
    if (typeof window.loadMorePosts === 'function') {
      console.log('Manually calling loadMorePosts...');
      window.loadMorePosts();
    } else {
      console.log('loadMorePosts function not found!');
    }
  });
  
  await page.waitForTimeout(3000);
  
  // Check post count after manual call
  const postCount = await page.locator('article.blog-post').count();
  console.log('Post count after manual loadMorePosts call:', postCount);
  
  // Now try scrolling
  console.log('Attempting scroll...');
  await page.evaluate(() => {
    console.log('Scrolling to bottom...');
    window.scrollTo(0, document.body.scrollHeight);
    console.log('After scroll - scrollY:', window.scrollY, 'scrollHeight:', document.body.scrollHeight);
  });
  
  await page.waitForTimeout(3000);
  
  const finalPostCount = await page.locator('article.blog-post').count();
  console.log('Final post count:', finalPostCount);
});