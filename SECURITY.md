# Security Policy

## Supported Versions

The following versions of slack-buddy-ai are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Security Features

### Built-in Security Measures

- **Token Validation**: All Slack tokens are validated and sanitized before use
- **Rate Limiting**: Exponential backoff prevents API abuse and respects Slack's rate limits
- **Input Sanitization**: All user inputs are validated and sanitized
- **Secure Defaults**: Configuration defaults prioritize security over convenience
- **No Token Logging**: Tokens are never logged or exposed in error messages

### Automated Security Scanning

This project employs multiple automated security scanning tools:

- **govulncheck**: Scans for known vulnerabilities in Go dependencies
- **gosec**: Static analysis for common security issues
- **nancy**: Dependency vulnerability scanning
- **staticcheck**: Advanced static analysis for bugs and security issues
- **Dependabot**: Automated dependency updates for security patches

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **DO NOT** Create a Public Issue

Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

### 2. Report Privately

Send an email to: **security@your-domain.com** (replace with actual contact)

Include the following information:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact assessment
- Any suggested fixes (if available)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Assessment**: Within 7 days
- **Fix Timeline**: 
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Within 60 days

### 4. Disclosure Policy

- We will acknowledge receipt of your vulnerability report
- We will provide an estimated timeline for fixes
- We will notify you when the vulnerability is fixed
- We will publicly disclose the vulnerability after a fix is available (coordinated disclosure)
- We will credit you for the discovery (unless you prefer to remain anonymous)

## Security Best Practices for Users

### Token Security

1. **Use Bot Tokens**: Always use bot tokens (`xoxb-`) rather than user tokens
2. **Minimal Permissions**: Grant only the required OAuth scopes:
   - `channels:read` - For channel detection
   - `chat:write` - For announcements (if needed)
3. **Environment Variables**: Store tokens in `.env` files or environment variables, never in code
4. **Token Rotation**: Regularly rotate Slack tokens as part of your security hygiene

### Installation Security

1. **Verify Downloads**: Always verify checksums of downloaded binaries
2. **Use Official Sources**: Download only from official GitHub releases
3. **Build from Source**: For maximum security, build from source code after review

### Operational Security

1. **Regular Updates**: Keep slack-buddy-ai updated to the latest version
2. **Monitor Logs**: Review application logs for suspicious activity
3. **Principle of Least Privilege**: Run with minimal required permissions
4. **Network Security**: Use appropriate network controls in production environments

### Configuration Security

```bash
# Secure .env file permissions
chmod 600 .env

# Example secure configuration
SLACK_TOKEN=xoxb-your-bot-token-here
```

## Security Scanning Results

This project undergoes regular security scanning:

- **Last Full Scan**: Updated with each release
- **Continuous Monitoring**: Automated daily scans via GitHub Actions
- **Dependency Updates**: Weekly automated security updates via Dependabot

## Responsible Disclosure Hall of Fame

We thank the following security researchers for responsibly disclosing vulnerabilities:

<!-- Add names here as vulnerabilities are reported and fixed -->
*No vulnerabilities reported yet.*

## Security Contacts

- **Security Issues**: security@your-domain.com
- **General Questions**: Create a GitHub issue (for non-security topics only)
- **Project Maintainer**: @astrostl

---

**Note**: This security policy is regularly reviewed and updated. Last updated: 2024-06-21.