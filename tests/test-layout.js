const puppeteer = require('puppeteer');

async function testEditorLayout() {
  console.log('🧪 Testing Editor Layout...\n');
  
  const browser = await puppeteer.launch({ 
    headless: false, // Show browser for debugging
    defaultViewport: { width: 1200, height: 800 }
  });
  
  const page = await browser.newPage();
  
  try {
    // Navigate to the test HTML file
    const testFile = 'file://' + __dirname + '/layout-test.html';
    console.log('📁 Loading:', testFile);
    await page.goto(testFile, { waitUntil: 'networkidle0' });
    
    // Wait for layout to be ready
    await page.waitForTimeout(1000);
    
    // Get layout information
    const layoutInfo = await page.evaluate(() => {
      const layout = document.querySelector('.editor-layout');
      const main = document.querySelector('.editor-main');
      const sidebar = document.querySelector('.editor-sidebar');
      
      if (!layout || !main || !sidebar) {
        return { error: 'Layout elements not found' };
      }
      
      const layoutRect = layout.getBoundingClientRect();
      const mainRect = main.getBoundingClientRect();
      const sidebarRect = sidebar.getBoundingClientRect();
      
      const layoutStyle = getComputedStyle(layout);
      const mainStyle = getComputedStyle(main);
      const sidebarStyle = getComputedStyle(sidebar);
      
      return {
        viewport: { width: window.innerWidth, height: window.innerHeight },
        layout: {
          rect: layoutRect,
          flexDirection: layoutStyle.flexDirection,
          display: layoutStyle.display
        },
        main: {
          rect: mainRect,
          flex: mainStyle.flex,
          order: mainStyle.order
        },
        sidebar: {
          rect: sidebarRect,
          flex: sidebarStyle.flex,
          order: sidebarStyle.order,
          position: sidebarStyle.position,
          width: sidebarStyle.width
        },
        isSideBySide: mainRect.x < sidebarRect.x,
        widthRatio: mainRect.width / sidebarRect.width
      };
    });
    
    if (layoutInfo.error) {
      console.log('❌ Error:', layoutInfo.error);
      return false;
    }
    
    console.log('📊 Layout Analysis:');
    console.log(`   Viewport: ${layoutInfo.viewport.width}x${layoutInfo.viewport.height}`);
    console.log(`   Layout display: ${layoutInfo.layout.display}`);
    console.log(`   Layout flex-direction: ${layoutInfo.layout.flexDirection}`);
    console.log(`   Main section: ${Math.round(layoutInfo.main.rect.width)}x${Math.round(layoutInfo.main.rect.height)} at (${Math.round(layoutInfo.main.rect.x)}, ${Math.round(layoutInfo.main.rect.y)})`);
    console.log(`   Main flex: ${layoutInfo.main.flex}, order: ${layoutInfo.main.order}`);
    console.log(`   Sidebar: ${Math.round(layoutInfo.sidebar.rect.width)}x${Math.round(layoutInfo.sidebar.rect.height)} at (${Math.round(layoutInfo.sidebar.rect.x)}, ${Math.round(layoutInfo.sidebar.rect.y)})`);
    console.log(`   Sidebar flex: ${layoutInfo.sidebar.flex}, order: ${layoutInfo.sidebar.order}`);
    console.log(`   Sidebar position: ${layoutInfo.sidebar.position}, width: ${layoutInfo.sidebar.width}`);
    console.log(`   Side by side: ${layoutInfo.isSideBySide}`);
    console.log(`   Width ratio: ${layoutInfo.widthRatio.toFixed(2)}:1\n`);
    
    // Verify layout is correct
    const isLayoutCorrect = 
      layoutInfo.layout.display === 'flex' &&
      layoutInfo.layout.flexDirection === 'row' &&
      layoutInfo.isSideBySide &&
      layoutInfo.sidebar.rect.width >= 300 &&
      layoutInfo.sidebar.rect.width <= 350 &&
      layoutInfo.widthRatio > 1.5;
    
    if (isLayoutCorrect) {
      console.log('✅ Layout test PASSED! Desktop layout is working correctly.');
    } else {
      console.log('❌ Layout test FAILED! Issues detected:');
      if (layoutInfo.layout.display !== 'flex') console.log('   - Layout is not using flexbox');
      if (layoutInfo.layout.flexDirection !== 'row') console.log('   - Layout is not side-by-side (should be row)');
      if (!layoutInfo.isSideBySide) console.log('   - Sidebar is not to the right of main content');
      if (layoutInfo.sidebar.rect.width < 300 || layoutInfo.sidebar.rect.width > 350) {
        console.log(`   - Sidebar width is ${Math.round(layoutInfo.sidebar.rect.width)}px (should be ~320px)`);
      }
      if (layoutInfo.widthRatio <= 1.5) {
        console.log(`   - Width ratio is ${layoutInfo.widthRatio.toFixed(2)}:1 (should be > 1.5:1)`);
      }
    }
    
    // Test mobile layout
    console.log('\n📱 Testing mobile layout...');
    await page.setViewport({ width: 375, height: 667 });
    await page.waitForTimeout(500);
    
    const mobileLayoutInfo = await page.evaluate(() => {
      const main = document.querySelector('.editor-main');
      const sidebar = document.querySelector('.editor-sidebar');
      
      const mainRect = main.getBoundingClientRect();
      const sidebarRect = sidebar.getBoundingClientRect();
      
      const layoutStyle = getComputedStyle(document.querySelector('.editor-layout'));
      
      return {
        viewport: { width: window.innerWidth, height: window.innerHeight },
        flexDirection: layoutStyle.flexDirection,
        main: { rect: mainRect },
        sidebar: { rect: sidebarRect },
        isStacked: sidebarRect.y > mainRect.y + mainRect.height - 100
      };
    });
    
    console.log(`   Mobile viewport: ${mobileLayoutInfo.viewport.width}x${mobileLayoutInfo.viewport.height}`);
    console.log(`   Layout flex-direction: ${mobileLayoutInfo.flexDirection}`);
    console.log(`   Main: ${Math.round(mobileLayoutInfo.main.rect.width)}x${Math.round(mobileLayoutInfo.main.rect.height)} at (${Math.round(mobileLayoutInfo.main.rect.x)}, ${Math.round(mobileLayoutInfo.main.rect.y)})`);
    console.log(`   Sidebar: ${Math.round(mobileLayoutInfo.sidebar.rect.width)}x${Math.round(mobileLayoutInfo.sidebar.rect.height)} at (${Math.round(mobileLayoutInfo.sidebar.rect.x)}, ${Math.round(mobileLayoutInfo.sidebar.rect.y)})`);
    console.log(`   Stacked layout: ${mobileLayoutInfo.isStacked}`);
    
    const isMobileCorrect = 
      mobileLayoutInfo.flexDirection === 'column' &&
      mobileLayoutInfo.isStacked;
      
    if (isMobileCorrect) {
      console.log('✅ Mobile layout test PASSED!');
    } else {
      console.log('❌ Mobile layout test FAILED!');
      if (mobileLayoutInfo.flexDirection !== 'column') console.log('   - Layout should be column on mobile');
      if (!mobileLayoutInfo.isStacked) console.log('   - Sidebar should be below main content on mobile');
    }
    
    // Keep browser open for manual inspection
    console.log('\n🔍 Browser will stay open for 10 seconds for manual inspection...');
    await page.waitForTimeout(10000);
    
    return isLayoutCorrect && isMobileCorrect;
    
  } catch (error) {
    console.error('❌ Test failed with error:', error);
    return false;
  } finally {
    await browser.close();
  }
}

// Run the test
testEditorLayout().then(success => {
  if (success) {
    console.log('\n🎉 All layout tests passed!');
    process.exit(0);
  } else {
    console.log('\n💥 Layout tests failed - needs more work.');
    process.exit(1);
  }
}).catch(error => {
  console.error('Test runner error:', error);
  process.exit(1);
});