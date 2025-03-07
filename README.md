# Furca

Keep your GitHub forks effortlessly fresh.

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

   ```
   GITHUB_TOKEN=your_github_token_here
   ```

You can also create a `.furca` file in your home directory with the same format.

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

## Example Output

```
Fetching forked repositories...
Found 5 forked repositories
Checking repository: awesome-project
Checking repository: cool-library
Checking repository: useful-tool
Checking repository: example-repo
Checking repository: test-project
✅ awesome-project is up to date with upstream
🔄 Successfully synced cool-library with upstream
❌ Error checking useful-tool: failed to compare commits: 404 Not Found
✅ example-repo is up to date with upstream
🔄 Successfully synced test-project with upstream
```

## Requirements

- Go 1.18 or higher
- GitHub personal access token with `repo` scope

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
