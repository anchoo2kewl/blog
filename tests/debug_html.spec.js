const { test, expect } = require('@playwright/test');

test('Debug HTML source and JavaScript errors', async ({ page }) => {
  // Capture console messages and errors
  page.on('console', msg => console.log('CONSOLE:', msg.type(), msg.text()));
  page.on('pageerror', error => console.log('PAGE ERROR:', error.message));
  
  await page.goto('http://localhost:22222');
  await page.waitForLoadState('networkidle');
  
  // Get the HTML source
  const htmlContent = await page.content();
  
  // Check if our script is in the HTML
  const scriptInHtml = htmlContent.includes('let offset = 5');
  console.log('Script found in HTML:', scriptInHtml);
  
  if (scriptInHtml) {
    const startIndex = htmlContent.indexOf('let offset = 5');
    const scriptSection = htmlContent.substring(startIndex, startIndex + 200);
    console.log('Script section:', scriptSection);
  }
  
  // Check for script tags
  const scriptTags = htmlContent.match(/<script[^>]*>[\s\S]*?<\/script>/gi) || [];
  console.log(`Found ${scriptTags.length} script tags`);
  
  // Check each script tag for our variables
  scriptTags.forEach((script, index) => {
    if (script.includes('offset') || script.includes('loadMorePosts')) {
      console.log(`Script ${index + 1} contains our infinite scroll code`);
      console.log('Script preview:', script.substring(0, 150) + '...');
    }
  });
  
  // Try a simple script injection test
  const testResult = await page.evaluate(() => {
    try {
      // Try to define a simple test variable
      window.testVar = 'working';
      return { 
        success: true, 
        testVar: window.testVar,
        errors: null 
      };
    } catch (error) {
      return { 
        success: false, 
        errors: error.message 
      };
    }
  });
  
  console.log('Script injection test:', testResult);
});