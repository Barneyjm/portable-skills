#!/usr/bin/env node

/**
 * Safely installs a portable-skills binary as a Claude Code skill.
 *
 * This script is a "good skill citizen":
 * - Only creates/modifies ~/.claude/skills/<skill-name>/
 * - Never touches other skills directories
 * - Creates parent directories safely with mkdir -p equivalent
 * - Warns before overwriting existing files
 *
 * Configuration is read from package.json:
 *   "portableSkill": {
 *     "name": "image",
 *     "binary": "ps-image"
 *   }
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');

// Read configuration from package.json
function getConfig() {
  const packageJsonPath = path.join(__dirname, '..', 'package.json');
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

  if (!packageJson.portableSkill) {
    console.error('Error: package.json missing "portableSkill" configuration');
    console.error('Expected: { "portableSkill": { "name": "...", "binary": "..." } }');
    process.exit(1);
  }

  return packageJson.portableSkill;
}

const config = getConfig();
const SKILL_NAME = config.name;
const BINARY_NAME = config.binary;

function getSkillsDir() {
  return path.join(os.homedir(), '.claude', 'skills');
}

function getOurSkillDir() {
  return path.join(getSkillsDir(), SKILL_NAME);
}

function getBinDir() {
  return path.join(__dirname, '..', 'bin');
}

async function confirm(question) {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer.toLowerCase() === 'y' || answer.toLowerCase() === 'yes');
    });
  });
}

async function main() {
  const binDir = getBinDir();
  const skillDir = getOurSkillDir();
  const skillsDir = getSkillsDir();

  // Check that we have the files to install
  const binaryExt = os.platform() === 'win32' ? '.exe' : '';
  const binaryPath = path.join(binDir, `${BINARY_NAME}${binaryExt}`);
  const skillMdPath = path.join(binDir, 'SKILL.md');

  if (!fs.existsSync(binaryPath)) {
    console.error(`Error: Binary not found at ${binaryPath}`);
    console.error('Run "npm install" first to download the binary.');
    process.exit(1);
  }

  if (!fs.existsSync(skillMdPath)) {
    console.error(`Error: SKILL.md not found at ${skillMdPath}`);
    console.error('Run "npm install" first to download SKILL.md.');
    process.exit(1);
  }

  console.log(`Installing ${SKILL_NAME} skill to ${skillDir}`);
  console.log('');

  // Check for existing skills directory and list what's there
  if (fs.existsSync(skillsDir)) {
    const existingSkills = fs.readdirSync(skillsDir).filter((name) => {
      const skillPath = path.join(skillsDir, name);
      return fs.statSync(skillPath).isDirectory();
    });

    if (existingSkills.length > 0) {
      console.log('Existing skills detected:');
      existingSkills.forEach((skill) => {
        const marker = skill === SKILL_NAME ? ' (will be updated)' : '';
        console.log(`  - ${skill}${marker}`);
      });
      console.log('');
    }
  }

  // Check if our skill directory already exists
  const destBinary = path.join(skillDir, `${BINARY_NAME}${binaryExt}`);
  const destSkillMd = path.join(skillDir, 'SKILL.md');

  if (fs.existsSync(skillDir)) {
    const existingFiles = [];
    if (fs.existsSync(destBinary)) existingFiles.push(BINARY_NAME);
    if (fs.existsSync(destSkillMd)) existingFiles.push('SKILL.md');

    if (existingFiles.length > 0) {
      console.log(`Warning: ${skillDir} already exists with:`);
      existingFiles.forEach((f) => console.log(`  - ${f}`));
      console.log('');

      const shouldContinue = await confirm('Overwrite existing files? (y/N) ');
      if (!shouldContinue) {
        console.log('Installation cancelled.');
        process.exit(0);
      }
      console.log('');
    }
  }

  // Create only our skill directory (mkdir -p equivalent)
  // This will create ~/.claude/skills/image/ without touching anything else
  fs.mkdirSync(skillDir, { recursive: true });

  // Copy binary
  console.log(`Copying ${BINARY_NAME}...`);
  fs.copyFileSync(binaryPath, destBinary);
  if (os.platform() !== 'win32') {
    fs.chmodSync(destBinary, 0o755);
  }

  // Copy SKILL.md
  console.log('Copying SKILL.md...');
  fs.copyFileSync(skillMdPath, destSkillMd);

  console.log('');
  console.log('✓ Installation complete!');
  console.log('');
  console.log('The skill is now available. You can invoke it with /image in Claude Code.');
  console.log('');
  console.log('Installed files:');
  console.log(`  ${destBinary}`);
  console.log(`  ${destSkillMd}`);
}

main().catch((err) => {
  console.error('Installation failed:', err.message);
  process.exit(1);
});
