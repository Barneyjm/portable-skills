#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, '..', 'bin');

// Clean up downloaded binary on uninstall
if (fs.existsSync(binDir)) {
  try {
    fs.rmSync(binDir, { recursive: true, force: true });
    console.log('Cleaned up ps-image binary');
  } catch (error) {
    // Ignore cleanup errors
  }
}
