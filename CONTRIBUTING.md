# Contributing to Slack Buddy AI

Thank you for your interest in contributing to Slack Buddy AI! We welcome contributions from the community.

## ⚠️ Project Status

This project is **"vibe coded"** (developed using generative AI tools) and maintained on a **best-effort basis** by volunteers. While we appreciate contributions, please understand:

- **No official support** - Community-driven maintenance
- **No guaranteed response times** - Volunteer availability varies
- **No SLA commitments** - Best-effort review and merge process

## 🤝 How to Contribute

### Issues

We welcome bug reports, feature requests, and questions:

1. **Search existing issues** first to avoid duplicates
2. **Use clear, descriptive titles** 
3. **Provide context**: OS, Go version, command used, expected vs actual behavior
4. **Include logs** (with tokens redacted) for debugging

**Issue Types Welcome:**
- 🐛 Bug reports
- ✨ Feature requests  
- 📚 Documentation improvements
- 🔧 Build/tooling enhancements
- 💡 Ideas and suggestions

### Pull Requests

We welcome pull requests for improvements:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/your-feature`)
3. **Make your changes** following the guidelines below
4. **Add/update tests** if applicable
5. **Run the test suite** (`make test`)
6. **Run linting** (`make lint`) 
7. **Submit a pull request**

## 🛠️ Development Guidelines

### Code Quality

- **Follow existing patterns** - Look at surrounding code for style
- **Add tests** for new functionality
- **Maintain coverage** - Don't decrease overall test coverage
- **Handle errors** properly with meaningful messages
- **Add logging** where appropriate (debug level for verbose info)

### Testing Requirements

```bash
# All tests must pass
make test

# Run with race detection
make test-race

# Check coverage (aim to maintain current levels)
make coverage

# Security and quality checks
make lint
```

### Commit Messages

- Use clear, descriptive commit messages
- Reference issues when applicable (`fixes #123`)
- Keep commits focused and atomic

### Documentation

- Update README.md if adding new features or changing usage
- Update CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/) format
- Add inline documentation for new public functions
- Update help text for new CLI commands/flags

## 🔒 Security

- **Never commit secrets** or tokens
- **Validate all inputs** especially user-provided data
- **Follow existing security patterns** in the codebase
- **Report security issues** following [SECURITY.md](SECURITY.md)

## 🧪 Testing Your Changes

### Local Testing

```bash
# Build and test locally
make dev

# Test with real Slack workspace (optional)
./slack-buddy health --verbose

# Run full CI-like checks
make ci
```

### Manual Testing Checklist

- [ ] Build succeeds on your platform
- [ ] All existing tests pass
- [ ] New functionality works as expected
- [ ] Help text is accurate and helpful
- [ ] Error messages are clear and actionable

## 📋 What We're Looking For

**High Value Contributions:**
- 🐛 Bug fixes with clear reproduction steps
- 🧪 Additional test coverage for edge cases
- 📚 Documentation improvements and clarifications
- ⚡ Performance improvements
- 🔧 Build/tooling enhancements

**Feature Additions:**
- Keep scope focused and manageable
- Align with project goals (Slack workspace management)
- Include comprehensive tests
- Update documentation

## 🚫 What to Avoid

- Large, monolithic changes without prior discussion
- Breaking changes without strong justification
- Features that significantly increase complexity
- Dependencies with restrictive licenses (GPL, etc.)
- Changes that reduce test coverage

## 💬 Getting Help

- **GitHub Issues** - For questions, bug reports, feature requests
- **Code Review** - We'll provide feedback on pull requests
- **Documentation** - Check README.md and inline help first

## 📜 License

By contributing, you agree that your contributions will be licensed under the same [MIT License](LICENSE) as the project.

## 🙏 Recognition

All contributors will be recognized in release notes and changelogs. Thank you for helping make Slack Buddy AI better!

---

**Remember:** This is a community project maintained by volunteers. We appreciate your patience and understanding.