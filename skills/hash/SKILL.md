---
name: portable-hash
description: Calculate and verify file checksums. Supports MD5, SHA1, SHA256, SHA512.
---

# Hash/Checksum Skill

Calculate and verify file checksums using industry-standard algorithms. No dependencies required.

## Available Commands

### calc
Calculate the hash of one or more files.

```bash
ps-hash calc <file> [file...] [--algorithm <algo>] [--json]
```

**Arguments:**
- `file` - Path to file(s) to hash. Use `-` for stdin.

**Flags:**
- `--algorithm, -a` - Hash algorithm: md5, sha1, sha256 (default), sha512
- `--json` - Output as JSON

**Examples:**
```bash
# Calculate SHA256 (default)
ps-hash calc document.pdf

# Calculate MD5
ps-hash calc file.zip -a md5

# Hash multiple files
ps-hash calc *.iso

# Hash from stdin
cat file.txt | ps-hash calc -

# JSON output
ps-hash calc file.bin --json
```

**Output format:**
```
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  document.pdf
```

### verify
Verify that a file matches an expected hash.

```bash
ps-hash verify <file> <expected-hash> [--algorithm <algo>]
```

**Arguments:**
- `file` - Path to file to verify
- `expected-hash` - Expected hash value

**Flags:**
- `--algorithm, -a` - Hash algorithm (default: sha256)

**Examples:**
```bash
# Verify SHA256
ps-hash verify download.iso abc123def456...

# Verify MD5
ps-hash verify file.zip d41d8cd98f00b204e9800998ecf8427e -a md5
```

**Exit codes:**
- `0` - Hash matches
- `1` - Hash does not match

## Supported Algorithms

| Algorithm | Output Length | Use Case |
|-----------|--------------|----------|
| MD5 | 32 chars | Legacy compatibility (not secure) |
| SHA1 | 40 chars | Legacy compatibility (not secure) |
| SHA256 | 64 chars | Recommended for most uses |
| SHA512 | 128 chars | Maximum security |

## Common Use Cases

1. **Verify downloads**: Check that a downloaded file matches the published checksum
2. **Detect changes**: Compare hashes to see if a file has been modified
3. **Generate checksums**: Create checksum files for distribution
4. **Data integrity**: Verify backups haven't been corrupted
