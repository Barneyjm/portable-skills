#!/usr/bin/env node

/**
 * Safely uninstalls ps-image from Claude Code skills.
 *
 * This script is a "good skill citizen":
 * - Only removes ~/.claude/skills/image/
 * - Never touches other skills directories
 * - Confirms before deleting
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');

const SKILL_NAME = 'image';

function getOurSkillDir() {
  return path.join(os.homedir(), '.claude', 'skills', SKILL_NAME);
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

function rmdir(dir) {
  if (fs.existsSync(dir)) {
    fs.readdirSync(dir).forEach((file) => {
      const curPath = path.join(dir, file);
      if (fs.lstatSync(curPath).isDirectory()) {
        rmdir(curPath);
      } else {
        fs.unlinkSync(curPath);
      }
    });
    fs.rmdirSync(dir);
  }
}

async function main() {
  const skillDir = getOurSkillDir();

  if (!fs.existsSync(skillDir)) {
    console.log(`Skill directory does not exist: ${skillDir}`);
    console.log('Nothing to uninstall.');
    process.exit(0);
  }

  // List what will be removed
  console.log(`Will remove: ${skillDir}`);
  console.log('');
  console.log('Contents:');
  fs.readdirSync(skillDir).forEach((file) => {
    console.log(`  - ${file}`);
  });
  console.log('');

  const shouldContinue = await confirm('Remove this skill? (y/N) ');
  if (!shouldContinue) {
    console.log('Uninstall cancelled.');
    process.exit(0);
  }

  // Remove only our directory
  rmdir(skillDir);

  console.log('');
  console.log('✓ Skill uninstalled.');
}

main().catch((err) => {
  console.error('Uninstall failed:', err.message);
  process.exit(1);
});
