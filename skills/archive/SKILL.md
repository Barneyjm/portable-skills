---
name: portable-archive
description: Create and extract zip archives. No dependencies required.
---

# Archive Skill

Create and extract zip archives without any system dependencies.

## Available Commands

### zip
Create a zip archive from files and directories.

```bash
ps-archive zip <file|dir> [file|dir...] [--output <name.zip>]
```

**Arguments:**
- `file|dir` - Files and/or directories to include

**Flags:**
- `--output, -o` - Output archive name (default: based on input)

**Examples:**
```bash
# Zip a single file
ps-archive zip document.pdf

# Zip with custom name
ps-archive zip photos/ -o vacation-photos.zip

# Zip multiple items
ps-archive zip file1.txt file2.txt folder/
```

### unzip
Extract files from a zip archive.

```bash
ps-archive unzip <archive.zip> [--output <dir>] [--overwrite]
```

**Arguments:**
- `archive.zip` - Archive to extract

**Flags:**
- `--output, -o` - Output directory (default: current directory)
- `--overwrite` - Overwrite existing files

**Examples:**
```bash
# Extract to current directory
ps-archive unzip backup.zip

# Extract to specific directory
ps-archive unzip photos.zip -o ~/Pictures/

# Overwrite existing files
ps-archive unzip update.zip --overwrite
```

### list
List contents of a zip archive.

```bash
ps-archive list <archive.zip> [--json]
```

**Flags:**
- `--json` - Output as JSON

**Examples:**
```bash
# List contents
ps-archive list backup.zip

# JSON output for parsing
ps-archive list backup.zip --json
```

**Output:**
```
   1.2 KB  document.pdf
   4.5 MB  photos/vacation.jpg
       -   photos/

3 entries, 4.7 MB (2.1 MB compressed)
```

## Security

The unzip command includes protection against "zip slip" attacks - it will reject archives containing paths that would extract outside the target directory.

## Common Use Cases

1. **Compress files for sharing**: Create zip archives for email or upload
2. **Extract downloads**: Unzip files downloaded from the web
3. **Backup folders**: Create compressed backups of directories
4. **Inspect archives**: List contents before extracting
