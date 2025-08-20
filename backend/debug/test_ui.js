const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

(async () => {
  const browser = await puppeteer.launch({ 
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  
  const page = await browser.newPage();
  
  try {
    console.log('Opening parse-test page...');
    await page.goto('http://localhost:6173/parse-test', { waitUntil: 'networkidle2' });
    
    // Read the match-test.txt file
    const matchTestPath = path.join(__dirname, 'match-test.txt');
    const matchLogs = fs.readFileSync(matchTestPath, 'utf8');
    console.log(`Loaded ${matchLogs.split('\n').length} lines from match-test.txt`);
    
    // Find the textarea and input the logs
    await page.waitForSelector('textarea');
    await page.evaluate((logs) => {
      document.querySelector('textarea').value = logs;
    }, matchLogs);
    
    // Click the Parse Logs button
    await page.click('button:has-text("Parse Logs")');
    
    // Wait for results
    await page.waitForSelector('table', { timeout: 10000 });
    await page.waitForTimeout(2000); // Give it time to render
    
    // Get the parsed results
    const results = await page.evaluate(() => {
      const stats = {};
      
      // Get summary stats
      const cards = document.querySelectorAll('.text-2xl.font-bold');
      if (cards.length >= 3) {
        stats.totalLines = cards[0].textContent;
        stats.parsedCount = cards[1].textContent;
        stats.failedCount = cards[2].textContent;
      }
      
      // Get event types from the table
      const eventTypes = {};
      const rows = document.querySelectorAll('tbody tr');
      
      rows.forEach(row => {
        const eventTypeBadge = row.querySelector('td:nth-child(3) .inline-flex');
        if (eventTypeBadge) {
          const eventType = eventTypeBadge.textContent.trim();
          eventTypes[eventType] = (eventTypes[eventType] || 0) + 1;
        }
      });
      
      return {
        stats,
        eventTypes,
        totalRows: rows.length
      };
    });
    
    console.log('\n=== Parse Test Results ===');
    console.log('Total Lines:', results.stats.totalLines);
    console.log('Parsed Count:', results.stats.parsedCount);
    console.log('Failed Count:', results.stats.failedCount);
    console.log('Table Rows:', results.totalRows);
    
    console.log('\n=== Event Type Distribution ===');
    const sortedTypes = Object.entries(results.eventTypes)
      .sort((a, b) => b[1] - a[1]);
    
    for (const [type, count] of sortedTypes) {
      console.log(`${type}: ${count}`);
    }
    
    // Check for "unrecognized" events
    const unrecognizedCount = sortedTypes
      .filter(([type]) => type.includes('unrecognized'))
      .reduce((sum, [, count]) => sum + count, 0);
    
    if (unrecognizedCount > 0) {
      console.log(`\n⚠️  Found ${unrecognizedCount} unrecognized events!`);
      
      // Get some examples
      const examples = await page.evaluate(() => {
        const rows = document.querySelectorAll('tbody tr');
        const unrecognizedExamples = [];
        
        rows.forEach(row => {
          const eventTypeBadge = row.querySelector('td:nth-child(3) .inline-flex');
          if (eventTypeBadge && eventTypeBadge.textContent.includes('unrecognized')) {
            const content = row.querySelector('td:nth-child(4) .font-mono').textContent;
            if (unrecognizedExamples.length < 5) {
              unrecognizedExamples.push({
                eventType: eventTypeBadge.textContent.trim(),
                content: content.substring(0, 100)
              });
            }
          }
        });
        
        return unrecognizedExamples;
      });
      
      console.log('\n=== Unrecognized Event Examples ===');
      examples.forEach((ex, i) => {
        console.log(`${i + 1}. ${ex.eventType}`);
        console.log(`   Content: ${ex.content}...`);
      });
    } else {
      console.log('\n✅ No unrecognized events found!');
    }
    
    // Take a screenshot for debugging
    await page.screenshot({ path: path.join(__dirname, 'parse-test-results.png'), fullPage: true });
    console.log('\nScreenshot saved to debug/parse-test-results.png');
    
  } catch (error) {
    console.error('Error during test:', error);
    // Take error screenshot
    await page.screenshot({ path: path.join(__dirname, 'parse-test-error.png'), fullPage: true });
  } finally {
    await browser.close();
  }
})();