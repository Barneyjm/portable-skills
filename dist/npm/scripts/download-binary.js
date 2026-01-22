#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const crypto = require('crypto');

const VERSION = '1.0.0';
const REPO = 'Barneyjm/portable-skills';
const BINARY_NAME = 'ps-image';

// Platform and architecture mapping
const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64',
};

function getBinaryName() {
  const platform = PLATFORM_MAP[os.platform()];
  const arch = ARCH_MAP[os.arch()];

  if (!platform || !arch) {
    console.error(`Unsupported platform: ${os.platform()}-${os.arch()}`);
    process.exit(1);
  }

  // Windows ARM64 not supported
  if (platform === 'windows' && arch === 'arm64') {
    console.error('Windows ARM64 is not supported. Please use x64.');
    process.exit(1);
  }

  const ext = platform === 'windows' ? '.exe' : '';
  return `${BINARY_NAME}-${platform}-${arch}${ext}`;
}

function getDownloadUrl(binaryName) {
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
}

function getChecksumsUrl() {
  return `https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt`;
}

function getSkillMdUrl() {
  return `https://github.com/${REPO}/releases/download/v${VERSION}/SKILL-image.md`;
}

async function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destPath);

    const request = (urlStr) => {
      https.get(urlStr, (response) => {
        // Handle redirects
        if (response.statusCode === 301 || response.statusCode === 302) {
          file.close();
          fs.unlinkSync(destPath);
          request(response.headers.location);
          return;
        }

        if (response.statusCode !== 200) {
          file.close();
          fs.unlinkSync(destPath);
          reject(new Error(`Failed to download: HTTP ${response.statusCode}`));
          return;
        }

        response.pipe(file);
        file.on('finish', () => {
          file.close();
          resolve();
        });
      }).on('error', (err) => {
        file.close();
        fs.unlinkSync(destPath);
        reject(err);
      });
    };

    request(url);
  });
}

async function fetchText(url) {
  return new Promise((resolve, reject) => {
    const request = (urlStr) => {
      https.get(urlStr, (response) => {
        // Handle redirects
        if (response.statusCode === 301 || response.statusCode === 302) {
          request(response.headers.location);
          return;
        }

        if (response.statusCode !== 200) {
          reject(new Error(`Failed to fetch: HTTP ${response.statusCode}`));
          return;
        }

        let data = '';
        response.on('data', (chunk) => { data += chunk; });
        response.on('end', () => resolve(data));
      }).on('error', reject);
    };

    request(url);
  });
}

function calculateSha256(filePath) {
  return new Promise((resolve, reject) => {
    const hash = crypto.createHash('sha256');
    const stream = fs.createReadStream(filePath);
    stream.on('error', reject);
    stream.on('data', (chunk) => hash.update(chunk));
    stream.on('end', () => resolve(hash.digest('hex')));
  });
}

async function verifyChecksum(filePath, binaryName, checksums) {
  const expectedLine = checksums.split('\n').find((line) => line.includes(binaryName));
  if (!expectedLine) {
    throw new Error(`Checksum not found for ${binaryName}`);
  }

  const expectedHash = expectedLine.split(/\s+/)[0];
  const actualHash = await calculateSha256(filePath);

  if (expectedHash !== actualHash) {
    throw new Error(
      `Checksum mismatch for ${binaryName}:\n` +
      `  Expected: ${expectedHash}\n` +
      `  Actual:   ${actualHash}`
    );
  }

  console.log(`✓ Checksum verified for ${binaryName}`);
}

async function main() {
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryName = getBinaryName();
  const destPath = path.join(binDir, os.platform() === 'win32' ? `${BINARY_NAME}.exe` : BINARY_NAME);

  // Create bin directory
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  console.log(`Downloading ${BINARY_NAME} v${VERSION} for ${os.platform()}-${os.arch()}...`);

  try {
    // Download binary
    const downloadUrl = getDownloadUrl(binaryName);
    console.log(`Fetching from: ${downloadUrl}`);
    await downloadFile(downloadUrl, destPath);

    // Download and verify checksums
    console.log('Verifying checksum...');
    const checksums = await fetchText(getChecksumsUrl());
    await verifyChecksum(destPath, binaryName, checksums);

    // Make executable (Unix only)
    if (os.platform() !== 'win32') {
      fs.chmodSync(destPath, 0o755);
    }

    console.log(`✓ Successfully installed ${BINARY_NAME} to ${destPath}`);

    // Download SKILL.md (required for skill functionality)
    const skillMdPath = path.join(binDir, 'SKILL.md');
    console.log('Downloading SKILL.md...');
    await downloadFile(getSkillMdUrl(), skillMdPath);
    console.log(`✓ Successfully installed SKILL.md to ${skillMdPath}`);
  } catch (error) {
    console.error(`Failed to install ${BINARY_NAME}:`, error.message);

    // Clean up partial download
    if (fs.existsSync(destPath)) {
      fs.unlinkSync(destPath);
    }

    process.exit(1);
  }
}

main();
