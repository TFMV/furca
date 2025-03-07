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
✅ translation-agent is up to date with upstream
❌ Error checking codon: failed to compare commits: GET https://api.github.com/repos/TFMV/codon/compare/exaloop%3Amaster...master: 404 Not Found []
✅ smallpond is up to date with upstream
✅ pgroll is up to date with upstream
✅ myduckserver is up to date with upstream
✅ stringtheory is up to date with upstream
✅ llvm-project is up to date with upstream
✅ go-capnp is up to date with upstream
❌ Error checking professional-services-data-validator: failed to compare commits: GET https://api.github.com/repos/TFMV/professional-services-data-validator/compare/GoogleCloudPlatform%3Amaster...master: 404 Not Found []
✅ gcp_data_utilities is up to date with upstream
✅ sheepda is up to date with upstream
✅ rclone is up to date with upstream
✅ automate-dv is up to date with upstream
✅ sklearn is up to date with upstream
✅ python-cluster is up to date with upstream
✅ sqlserver2pgsql is up to date with upstream
✅ pattern is up to date with upstream
✅ bodkin is up to date with upstream
✅ spanner-migration-tool is up to date with upstream
✅ gocql is up to date with upstream
✅ act is up to date with upstream
✅ gazette-core is up to date with upstream
✅ resume-cli is up to date with upstream
✅ trillian is up to date with upstream
✅ urho-samples is up to date with upstream
✅ SpatialSearch is up to date with upstream
✅ petl is up to date with upstream
✅ resume-schema is up to date with upstream
✅ genAI is up to date with upstream
✅ ora2pg is up to date with upstream
✅ simdjson-go is up to date with upstream
✅ slim is up to date with upstream
🔄 Successfully synced weaviate with upstream
❌ Failed to sync pdf2json: failed to sync repository: failed to execute request: POST https://api.github.com/repos/TFMV/pdf2json/merge-upstream: 409 There are merge conflicts []
```

## Requirements

- Go 1.18 or higher
- GitHub personal access token with `repo` scope

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
