# Furca

Keep your GitHub forks effortlessly fresh.

[![Go Reference](https://pkg.go.dev/badge/github.com/TFMV/furca.svg)](https://pkg.go.dev/github.com/TFMV/furca)
[![Go Report Card](https://goreportcard.com/badge/github.com/TFMV/furca)](https://goreportcard.com/report/github.com/TFMV/furca)
[![Go 1.24](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/github/license/TFMV/furca)](LICENSE)

## Table of Contents

- [Furca](#furca)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Installation](#installation)
    - [From Source](#from-source)
    - [Using Go Install](#using-go-install)
  - [Configuration](#configuration)
    - [Additional Configuration Options](#additional-configuration-options)
  - [Usage](#usage)
    - [Sync Command](#sync-command)
    - [Advanced Options](#advanced-options)
      - [Dry Run Mode](#dry-run-mode)
      - [JSON Output](#json-output)
      - [Retry Configuration](#retry-configuration)
      - [Log Level](#log-level)
  - [Example Output](#example-output)
  - [Requirements](#requirements)
  - [License](#license)

## Overview

Furca is a command-line tool built in Go designed to automate the synchronization of forked GitHub repositories with their upstream sources. It simplifies the developer experience by automatically fetching repository information, determining if forks are behind their upstream repositories, and synchronizing them accordingly when executed.

## Features

- Automatically detects all your forked repositories
- Identifies which forks are behind their upstream repositories
- Synchronizes forks with upstream changes
- Concurrent processing for efficiency
- Clear, structured output showing sync status

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/TFMV/furca.git
cd furca

# Build the binary
go build -o furca

# Move to a directory in your PATH (optional)
mv furca /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/TFMV/furca@latest
```

## Configuration

Furca requires a GitHub personal access token with the `repo` scope to access your repositories. You can provide this token in one of two ways:

1. Environment variable:

   ```bash
   export GITHUB_TOKEN=your_github_token_here
   ```

2. `.env` file in the current directory:

   ```bash
   GITHUB_TOKEN=your_github_token_here
   ```

You can also create a `.furca` file in your home directory with the same format.

### Additional Configuration Options

You can configure the following options either via command-line flags or in your `.env` file:

| Environment Variable | Command-line Flag | Description | Default |
|----------------------|-------------------|-------------|---------|
| `GITHUB_TOKEN` | - | GitHub personal access token | (required) |
| `LOG_LEVEL` | - | Logging verbosity (debug, info, warn, error) | info |
| `DRY_RUN` | `--dry-run` | Preview changes without syncing | false |
| `JSON_OUTPUT` | `--json` | Output results in JSON format | false |
| `MAX_RETRIES` | `--max-retries` | Maximum retry attempts for API operations | 2 |
| `RETRY_DELAY` | `--retry-delay` | Delay in seconds between retries | 3 |

Example `.env` file:

```bash
GITHUB_TOKEN=your_github_token_here
LOG_LEVEL=debug
DRY_RUN=true
JSON_OUTPUT=true
MAX_RETRIES=3
RETRY_DELAY=5
```

Command-line flags take precedence over environment variables.

## Usage

### Sync Command

The primary command is `sync`, which synchronizes all your forked repositories with their upstream sources:

```bash
furca sync
```

This will:

1. Fetch all your forked repositories
2. Check which ones are behind their upstream
3. Synchronize the ones that are behind
4. Display the results

### Advanced Options

#### Dry Run Mode

Preview which repositories would be synced without making any changes:

```bash
furca sync --dry-run
```

Example output:

```bash
[DRY-RUN] ✅ awesome-project is up to date with upstream
[DRY-RUN] 🔄 Would sync cool-library (behind by 5 commits)
```

#### JSON Output

Get structured JSON output for integration with other tools:

```bash
furca sync --json
```

Example output:

```json
{
  "synced": ["weaviate", "duckdb-wasm"],
  "up_to_date": ["pattern", "simdjson-go"],
  "errors": {
    "codon": "failed to compare commits: 404 Not Found"
  },
  "timestamp": "2025-03-07T16:30:00Z"
}
```

#### Retry Configuration

Configure retry behavior for API operations:

```bash
furca sync --max-retries=3 --retry-delay=5
```

#### Log Level

Control the verbosity of logging:

```bash
# In .env file or environment variable
LOG_LEVEL=debug  # Options: debug, info, warn, error, dpanic, panic, fatal
```

## Example Output

```bash
Fetching forked repositories...
Found 5 forked repositories
✅ awesome-project is up to date with upstream
🔄 Successfully synced cool-library with upstream (was behind by 2 commits)
❌ Error checking useful-tool: failed to compare commits: 404 Not Found
✅ example-repo is up to date with upstream
🔄 Successfully synced test-project with upstream (was behind by 5 commits)

📊 Summary:
🔄 Synced repositories: 2
✅ Up-to-date repositories: 2
❌ Errors encountered: 1

See logs for details.
```

## Requirements

- Go 1.18 or higher
- GitHub personal access token with `repo` scope

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
