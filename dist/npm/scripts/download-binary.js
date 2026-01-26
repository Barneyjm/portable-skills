#!/usr/bin/env node

/**
 * Downloads all platform-specific binaries for portable-skills.
 *
 * SKILL.md is embedded in each binary, so we only need to download
 * the binaries. Run `<binary> install-skill` after installation
 * to set up each Claude Code skill.
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const crypto = require('crypto');

// Read configuration from package.json
function getConfig() {
  const packageJsonPath = path.join(__dirname, '..', 'package.json');
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

  return {
    version: packageJson.version,
    repo: packageJson.repository.url.replace('git+https://github.com/', '').replace('.git', ''),
    binaries: packageJson.portableSkills?.binaries || ['ps-image'],
  };
}

const config = getConfig();
const VERSION = config.version;
const REPO = config.repo;
const BINARIES = config.binaries;

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

function getPlatformArch() {
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

  return { platform, arch };
}

function getRemoteBinaryName(binaryName, platform, arch) {
  const ext = platform === 'windows' ? '.exe' : '';
  return `${binaryName}-${platform}-${arch}${ext}`;
}

function getLocalBinaryName(binaryName) {
  const ext = os.platform() === 'win32' ? '.exe' : '';
  return `${binaryName}${ext}`;
}

function getDownloadUrl(remoteBinaryName) {
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${remoteBinaryName}`;
}

function getChecksumsUrl() {
  return `https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt`;
}

async function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destPath);

    const request = (urlStr) => {
      https.get(urlStr, (response) => {
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
        if (fs.existsSync(destPath)) fs.unlinkSync(destPath);
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

async function verifyChecksum(filePath, remoteBinaryName, checksums) {
  const expectedLine = checksums.split('\n').find((line) => line.includes(remoteBinaryName));
  if (!expectedLine) {
    throw new Error(`Checksum not found for ${remoteBinaryName}`);
  }

  const expectedHash = expectedLine.split(/\s+/)[0];
  const actualHash = await calculateSha256(filePath);

  if (expectedHash !== actualHash) {
    throw new Error(
      `Checksum mismatch for ${remoteBinaryName}:\n` +
      `  Expected: ${expectedHash}\n` +
      `  Actual:   ${actualHash}`
    );
  }
}

async function downloadBinary(binaryName, binDir, platform, arch, checksums) {
  const remoteName = getRemoteBinaryName(binaryName, platform, arch);
  const localName = getLocalBinaryName(binaryName);
  const destPath = path.join(binDir, localName);
  const downloadUrl = getDownloadUrl(remoteName);

  console.log(`  Downloading ${binaryName}...`);
  await downloadFile(downloadUrl, destPath);
  await verifyChecksum(destPath, remoteName, checksums);

  if (os.platform() !== 'win32') {
    fs.chmodSync(destPath, 0o755);
  }

  console.log(`  ✓ ${binaryName}`);
}

async function main() {
  const binDir = path.join(__dirname, '..', 'bin');
  const { platform, arch } = getPlatformArch();

  // Create bin directory
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  console.log(`Downloading portable-skills v${VERSION} for ${os.platform()}-${os.arch()}...`);
  console.log(`Skills: ${BINARIES.join(', ')}`);
  console.log('');

  try {
    // Fetch checksums once for all binaries
    const checksums = await fetchText(getChecksumsUrl());

    // Download all binaries
    for (const binary of BINARIES) {
      await downloadBinary(binary, binDir, platform, arch, checksums);
    }

    console.log('');
    console.log(`✓ Successfully installed ${BINARIES.length} skills`);
    console.log('');
    console.log('To install as Claude Code skills, run:');
    for (const binary of BINARIES) {
      console.log(`  ${binary} install-skill`);
    }
    console.log('');
    console.log('Or install all at once:');
    console.log(`  ${BINARIES.map(b => `${b} install-skill`).join(' && ')}`);
  } catch (error) {
    console.error('Failed to install portable-skills:', error.message);
    process.exit(1);
  }
}

main();
