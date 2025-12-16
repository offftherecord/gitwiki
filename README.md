# Gitwiki

A security research tool for discovering publicly writable GitHub wikis across organizations and user accounts. Writable wikis can be exploited for social engineering and phishing attacks, as adversaries can modify or add wiki pages with malicious instructions that users may trust and follow.

## Overview

Gitwiki scans GitHub organizations and user accounts to identify repositories with publicly writable wikis. Since users typically trust content in organization repositories, writable wikis present a security risk that should be identified and remediated by security teams.

## Features

- **Auto-detection**: Automatically detects whether an account is an organization or user
- **Explicit targeting**: Support for `org:` and `user:` prefixes to specify account type
- **Authentication**: Optional GitHub token support for higher rate limits (via `GITHUB_TOKEN` environment variable)
- **Rate limiting**: Automatic rate limit detection and waiting
- **Public repositories only**: Filters for public repositories with enabled wikis
- **Write verification**: Tests actual wiki writeability, not just permissions

## Installation

Gitwiki requires **Go 1.21** or higher. Install using:

```bash
go install github.com/offftherecord/gitwiki@latest
```

## Usage

### Basic Usage

Scan an organization (auto-detect):
```bash
gitwiki UrbanCompass
```

Scan a user account (auto-detect):
```bash
gitwiki offftherecord
```

### Explicit Account Type

Specify organization explicitly:
```bash
gitwiki org:UrbanCompass
```

Specify user explicitly:
```bash
gitwiki user:offftherecord
```

### Batch Scanning

Scan multiple accounts via stdin:
```bash
echo "UrbanCompass" | gitwiki
cat accounts.txt | gitwiki
```

### Authentication

For higher rate limits, set a GitHub personal access token:
```bash
export GITHUB_TOKEN=your_token_here
gitwiki UrbanCompass
```

## Output

Gitwiki reports two types of writable wikis:

- **Writable**: Wiki exists with pages, but allows public write access
  ```
  Writable: repo-name, URL: https://github.com/org/repo-name/wiki/notrealpage
  ```

- **Writable-Firstpage**: Wiki is enabled but has no pages yet (completely empty)
  ```
  Writable-Firstpage: repo-name, URL: https://github.com/org/repo-name/wiki
  ```

## How It Works

1. Fetches all public repositories for the specified organization or user
2. Filters for repositories with wikis enabled
3. Tests wiki writeability by:
   - Checking if the wiki has no first page (indicates write access)
   - Attempting to access a non-existent page (write-protected wikis redirect to login)
4. Reports only wikis that are publicly writable

## Security Considerations

This tool is intended for:
- **Offensive security**: Identifying potential social engineering vectors
- **Defensive security**: Auditing organizational repositories for misconfigured wikis
- **Security research**: Understanding the scope of publicly writable wikis on GitHub

Organizations should regularly audit their repositories to ensure wikis are either disabled or properly configured with write restrictions.

## License

This project is licensed under the MIT License. See the [LICENSE.md](LICENSE.md) file for details.
