# Release Process

This document covers the complete release process for slack-butler, including both Go module releases (for `go install`) and Homebrew releases.

**⚠️ CRITICAL: STOP IMMEDIATELY IF ANY STEP FAILS ⚠️**

If any command returns an error or fails quality checks, STOP the release process immediately. Fix the issue, commit the fix, and restart from the beginning.

---

## Table of Contents

1. [Pre-Release Requirements](#pre-release-requirements)
2. [Go Module Release Process](#go-module-release-process)
3. [Homebrew Release Process](#homebrew-release-process)
4. [Docker Release Process](#docker-release-process)
5. [Common Issues and Solutions](#common-issues-and-solutions)
6. [Post-Release Verification](#post-release-verification)

---

## Pre-Release Requirements

Before starting ANY release process, ensure these requirements are met:

### 1. Working Directory is Clean

```bash
git status
```

- **MUST** show no uncommitted changes
- The version will show as "-dirty" if there are uncommitted changes, breaking the release
- If any uncommitted changes exist, commit or stash them

### 2. All Quality Gates Pass

**MANDATORY** - Run this exact sequence and ALL must pass:

```bash
make clean && make deps && make quality && make test && make coverage && make build
```

Quality gate requirements:
- ✅ **Complete Test Suite**: ALL tests must pass with 100% success rate
- ✅ **Quality Checks**: ALL linting, security, and complexity checks must pass
- ✅ **Coverage Validation**: Test coverage remains comprehensive
- ✅ **Build Verification**: Binary must compile successfully
- ✅ **Race Condition Testing**: All tests must pass with race detection enabled

**ABSOLUTE PROHIBITIONS:**
- ❌ **NEVER** push releases with failing tests
- ❌ **NEVER** push releases with linting errors
- ❌ **NEVER** push releases with security issues (gosec failures)
- ❌ **NEVER** push releases with build failures
- ❌ **NEVER** skip quality checks "just this once"

### 3. Documentation is Updated

Update all documentation BEFORE creating the release tag:

#### CHANGELOG.md
- Add new version section with date
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Use `-beta`, `-alpha` suffixes for pre-releases
- Document all breaking changes, new features, and bug fixes

#### README.md
- Update usage examples if commands changed
- Add new features to feature list
- Update installation instructions if needed
- Update version badge
- Ensure roadmap reflects current plans

#### CLAUDE.md
- Record version changes and release status
- Update project structure if modified
- Add new development commands or processes

**Commit and push documentation updates:**

```bash
git add CHANGELOG.md README.md CLAUDE.md
git commit -m "Update documentation for v1.X.X release"
git push origin main
```

---

## Go Module Release Process

This process makes the release available via `go install github.com/astrostl/slack-butler@latest`.

### Step 1: Create and Push Version Tag

```bash
# Create the version tag
git tag v1.X.X

# Push the tag to GitHub
git push origin main --tags
```

**That's it!** For Go module releases, the git tag is all that's needed. Users can now install via:

```bash
go install github.com/astrostl/slack-butler@v1.X.X
go install github.com/astrostl/slack-butler@latest
```

### Step 2: Verify Go Install Works

Test that the release is installable:

```bash
# Remove existing installation (if any)
rm -f ~/go/bin/slack-butler

# Install from the new tag
go install github.com/astrostl/slack-butler@v1.X.X

# Verify version
slack-butler --version
```

Should show: `slack-butler v1.X.X` (clean version, no suffixes)

**If you want to do a Homebrew release as well, continue to the next section.**

---

## Homebrew Release Process

This process makes the release available via `brew install astrostl/slack-butler/slack-butler`.

**IMPORTANT**: The Go module release (git tag) MUST be completed first!

### Step 1: Verify Git Tag Exists

Ensure the git tag was created in the Go release process:

```bash
git tag | grep v1.X.X
```

If the tag doesn't exist, go back to the [Go Module Release Process](#go-module-release-process).

### Step 2: Build Homebrew Assets

Build macOS binaries and generate checksums:

```bash
# Clean any previous builds
rm -rf dist

# Build binaries, package them, and generate checksums
make build-macos-binaries package-macos-binaries generate-macos-checksums
```

**CRITICAL:** The working directory MUST be clean. If you've made any commits since creating the tag, the binaries will have a dirty version and you MUST move the tag:

```bash
git tag -f v1.X.X && git push -f origin main --tags
```

### Step 3: Verify Binary Versions

Before creating the GitHub release, verify the binaries have the correct clean version:

```bash
# Check AMD64 version
./dist/slack-butler-darwin-amd64 --version

# Check ARM64 version
./dist/slack-butler-darwin-arm64 --version
```

Both should show exactly `slack-butler v1.X.X` with NO suffixes like `-dirty` or `-N-gHASH`.

**If versions are wrong:**
1. You made commits after creating the tag
2. Move the tag: `git tag -f v1.X.X && git push -f origin main --tags`
3. Rebuild: `rm -rf dist && make build-macos-binaries package-macos-binaries generate-macos-checksums`

### Step 4: Record Checksums for Formula

**CRITICAL:** Save these checksums NOW - you'll need them in Step 6:

```bash
# Display and save checksums
cat dist/checksums.txt
```

**Copy these SHA256 values to a text file or leave this terminal window open.** The Homebrew formula MUST use these EXACT checksums that match the GitHub release assets.

### Step 5: Create GitHub Release

Create the GitHub release with the generated assets:

```bash
gh release create v1.X.X \
  --title "v1.X.X - Release Title" \
  --notes "## Added
- Feature 1
- Feature 2

## Changed
- Change 1
- Change 2

## Fixed
- Bug fix 1

🍺 **Homebrew Installation:**
\`\`\`bash
brew install astrostl/slack-butler/slack-butler
\`\`\`

📦 **Go Install:**
\`\`\`bash
go install github.com/astrostl/slack-butler@latest
\`\`\`

🔨 **Build from Source:**
\`\`\`bash
git clone https://github.com/astrostl/slack-butler.git
cd slack-butler
git checkout v1.X.X
go build -o slack-butler
\`\`\`

Generated with [Claude Code](https://claude.com/claude-code)" \
  dist/slack-butler-v1.X.X-darwin-amd64.tar.gz \
  dist/slack-butler-v1.X.X-darwin-arm64.tar.gz \
  dist/checksums.txt
```

**IMPORTANT:** GitHub may take a few seconds to process the assets. Wait until the release page loads before proceeding.

### Step 6: Update Homebrew Formula with Correct Checksums

**CRITICAL:** Manually update the formula with the checksums you saved in Step 4.

**DO NOT run `make update-homebrew-formula`** - it will rebuild binaries with different checksums than what was uploaded to GitHub.

```bash
# Edit the formula manually
nano Formula/slack-butler.rb
# OR use your preferred editor
code Formula/slack-butler.rb
```

Update these lines with the checksums from Step 4:
- Line 4: Version number (remove 'v' prefix: `1.X.X`)
- Line 8: ARM64 URL (for `darwin-arm64.tar.gz`)
- Line 9: ARM64 sha256
- Line 11: AMD64 URL (for `darwin-amd64.tar.gz`)
- Line 12: AMD64 sha256

**Verify the checksums match GitHub release:**

```bash
# Download and verify ARM64 checksum from GitHub
curl -sL https://github.com/astrostl/slack-butler/releases/download/v1.X.X/slack-butler-v1.X.X-darwin-arm64.tar.gz | shasum -a 256

# Download and verify AMD64 checksum from GitHub
curl -sL https://github.com/astrostl/slack-butler/releases/download/v1.X.X/slack-butler-v1.X.X-darwin-amd64.tar.gz | shasum -a 256

# Compare with formula
grep sha256 Formula/slack-butler.rb
```

All three sources (Step 4 checksums, GitHub download checksums, formula checksums) MUST match exactly.

### Step 7: Commit and Push Formula

Commit the formula with correct checksums:

```bash
git add Formula/slack-butler.rb
git commit -m "Update Homebrew formula for v1.X.X with correct SHA256 checksums"
git push origin main
```

**CRITICAL:** Do NOT move the release tag after this point. The GitHub release already has the correct binaries.

### Step 8: Verify Homebrew Installation

Test that the formula works correctly:

```bash
# Update Homebrew to get latest formula
brew update

# Clear any cached downloads (in case of previous failed attempts)
rm -f ~/Library/Caches/Homebrew/downloads/*slack-butler-v1.X.X*

# Install or upgrade slack-butler
brew upgrade slack-butler
# OR if not installed yet:
brew install astrostl/slack-butler/slack-butler

# Verify version
slack-butler --version
```

Should show: `slack-butler v1.X.X` (clean version, no suffixes)

**If you get a SHA256 mismatch error:**
1. The checksums in the formula don't match the GitHub release assets
2. Download one of the release assets manually and verify its checksum
3. Update the formula with the correct checksum and repeat Step 7

---

## Docker Release Process

This process publishes multi-platform Docker images to Docker Hub at `astrostl/slack-butler`.

**IMPORTANT**: The Go module release (git tag) MUST be completed first!

### Step 1: Verify Git Tag Exists

Ensure the git tag was created in the Go release process:

```bash
git tag | grep v1.X.X
```

If the tag doesn't exist, go back to the [Go Module Release Process](#go-module-release-process).

### Step 2: Build Docker Image Locally

Build and test the Docker image locally first:

```bash
# Build Docker image with version tag
make docker-build

# Test the image works correctly
docker run astrostl/slack-butler:v1.X.X --version
docker run astrostl/slack-butler:v1.X.X --help
```

The version should show exactly `slack-butler v1.X.X` with no suffixes.

### Step 3: Login to Docker Hub

Authenticate with Docker Hub:

```bash
docker login
# OR for nerdctl:
nerdctl login
```

Enter your Docker Hub credentials when prompted.

### Step 4: Build and Push Multi-Platform Images

Build and push images for both linux/amd64 and linux/arm64:

```bash
make docker-push
```

This will:
1. Build AMD64 image and push with tags: `latest-amd64` and `v1.X.X-amd64`
2. Build ARM64 image and push with tags: `latest-arm64` and `v1.X.X-arm64`
3. Create multi-platform manifests for `latest` and `v1.X.X` tags

**Alternative:** For single-platform testing only:
```bash
make docker-push-single
```

### Step 5: Verify Docker Hub Publication

Check that images are available on Docker Hub:

```bash
# Pull and test the latest tag
docker pull astrostl/slack-butler:latest
docker run astrostl/slack-butler:latest --version

# Pull and test the version tag
docker pull astrostl/slack-butler:v1.X.X
docker run astrostl/slack-butler:v1.X.X --version
```

You can also verify on Docker Hub web interface:
- https://hub.docker.com/r/astrostl/slack-butler/tags

Both `latest` and `v1.X.X` should show support for linux/amd64 and linux/arm64.

### Step 6: Test Multi-Platform Images

Test on different architectures if available:

```bash
# Test on AMD64 system
docker run --platform linux/amd64 astrostl/slack-butler:latest --version

# Test on ARM64 system (Mac Silicon, ARM servers, etc.)
docker run --platform linux/arm64 astrostl/slack-butler:latest --version
```

---

## Common Issues and Solutions

### Issue: Binary Version Shows "v1.X.X-1-gHASH"

**Cause:** Commits were made after creating the release tag.

**Solution:**
1. Move the tag: `git tag -f v1.X.X && git push -f origin main --tags`
2. For Homebrew: Rebuild assets: `rm -rf dist && make build-macos-binaries package-macos-binaries generate-macos-checksums`
3. For Homebrew: Delete and recreate the GitHub release with the new binaries

### Issue: Binary Version Shows "v1.X.X-dirty"

**Cause:** Uncommitted changes in working directory.

**Solution:**
1. Commit or stash all changes
2. Move the tag: `git tag -f v1.X.X && git push -f origin main --tags`
3. For Homebrew: Rebuild everything from Homebrew Step 2 onward

### Issue: Homebrew SHA256 Mismatch

**Cause:** The checksums in the formula don't match the GitHub release assets.

**Root Cause:** Running `make update-homebrew-formula` in Step 6 rebuilds binaries with different checksums than what was uploaded to GitHub.

**Solution:**
1. Download a release asset from GitHub and verify its checksum:
   ```bash
   curl -sL https://github.com/astrostl/slack-butler/releases/download/v1.X.X/slack-butler-v1.X.X-darwin-arm64.tar.gz | shasum -a 256
   ```
2. Manually edit `Formula/slack-butler.rb` with the correct checksum
3. DO NOT use `make update-homebrew-formula` - it rebuilds binaries
4. Commit and push the formula fix
5. Clear Homebrew cache: `rm -f ~/Library/Caches/Homebrew/downloads/*slack-butler*`
6. Test: `brew upgrade slack-butler`

### Issue: Quality Gates Fail

**Cause:** Code doesn't meet quality standards.

**Solution:**
1. NEVER skip quality checks or push releases with failures
2. Fix all linting errors, test failures, and security issues
3. Run `make clean && make deps && make quality && make test && make coverage && make build`
4. Only proceed when ALL checks pass
5. Commit fixes and restart from Pre-Release Requirements

### Issue: `go install` Shows "dev" or Wrong Version

**Cause:** The version information is set to "dev" by default in main.go and only gets updated via ldflags during `make build`. When users run `go install`, the ldflags aren't applied, so the version shows as "dev".

**Root Cause:** Go's `go install` command doesn't use our Makefile's ldflags, so the version variables in main.go remain at their default values.

**Solution (IMPLEMENTED):** Use Go's `runtime/debug.ReadBuildInfo()` to automatically extract the version from the module information when ldflags aren't provided. This ensures:
- `make build` uses ldflags for full version info (version + build time + commit)
- `go install` uses the module version from build info (version only)
- Both methods show the correct version number

**Code Fix:**
```go
// In main.go
import "runtime/debug"

func main() {
    version := Version
    if version == "dev" {
        if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
            version = info.Main.Version
        }
    }
    cmd.Execute(version, BuildTime, GitCommit)
}
```

**Verification:**
```bash
# After fix is merged and tagged
go install github.com/astrostl/slack-butler@v1.X.X
slack-butler --version  # Should show: slack-butler version v1.X.X
```

### Issue: `go install` Shows Old Version (Proxy Cache)

**Cause:** Go module proxy cache hasn't refreshed yet after a new release.

**Solution:**
- Wait 5-10 minutes for the proxy cache to refresh
- Force refresh: `GOPROXY=direct go install github.com/astrostl/slack-butler@v1.X.X`

### Issue: Need to Inspect Docker Image Contents

**Purpose:** Verify the Docker image contains only the expected files (binary + CA certificates).

**Methods to inspect the image:**

1. **View Docker history (recommended):**
   ```bash
   docker history astrostl/slack-butler:v1.X.X --no-trunc --format "{{.CreatedBy}}"
   ```
   Shows what was copied into the image.

2. **Inspect image configuration:**
   ```bash
   docker inspect astrostl/slack-butler:v1.X.X --format='{{json .Config}}' | jq
   ```
   Shows entrypoint, command, and other config details.

3. **Review the Dockerfile:**
   The Dockerfile explicitly shows what's included:
   - `/slack-butler` - The compiled binary
   - `/etc/ssl/certs/ca-certificates.crt` - CA certificates for HTTPS

**Expected Contents:**
```
/ (root)
├── slack-butler                           # Main executable
└── etc/ssl/certs/ca-certificates.crt     # CA certs for Slack API
```

**Note:** The image is based on `scratch` (empty base), so it has NO shell, NO OS utilities, and NO package manager. This is intentional for security and minimal size (~3 MB total).

---

## Post-Release Verification

After completing the release process, verify:

### For Go Module Releases:
1. ✅ Git tag exists: `git tag | grep v1.X.X`
2. ✅ Tag pushed to remote: `git ls-remote --tags origin | grep v1.X.X`
3. ✅ `go install` works: `go install github.com/astrostl/slack-butler@v1.X.X`
4. ✅ Installed binary shows clean version: `slack-butler --version`

### For Homebrew Releases (in addition to above):
1. ✅ GitHub release exists with binaries and checksums
2. ✅ Homebrew upgrade/install succeeds without checksum errors
3. ✅ Installed binary shows clean version: `slack-butler --version`
4. ✅ Formula in repository has correct checksums matching GitHub release assets
5. ✅ All quality gates passed before release

### For Docker Releases (in addition to Go releases):
1. ✅ Docker image builds successfully: `make docker-build`
2. ✅ Local test passes: `docker run astrostl/slack-butler:v1.X.X --version`
3. ✅ Multi-platform images pushed to Docker Hub: `make docker-push`
4. ✅ Images available on Docker Hub with `latest` and `v1.X.X` tags
5. ✅ Both linux/amd64 and linux/arm64 platforms supported
6. ✅ Pull and run tests pass: `docker pull astrostl/slack-butler:latest`

---

## Release Checklist

### Pre-Release Checklist (Required for Both Go and Homebrew)

- [ ] Working directory is clean (`git status`)
- [ ] Quality gates passed (`make clean && make deps && make quality && make test && make coverage && make build`)
- [ ] Documentation updated (CHANGELOG.md, README.md, CLAUDE.md)
- [ ] Documentation committed and pushed

### Go Module Release Checklist

- [ ] Release tag created (`git tag v1.X.X`)
- [ ] Release tag pushed (`git push origin main --tags`)
- [ ] `go install` tested and version verified

### Homebrew Release Checklist (Optional, after Go Release)

- [ ] Git tag verified to exist
- [ ] Homebrew binaries built (`rm -rf dist && make build-macos-binaries package-macos-binaries generate-macos-checksums`)
- [ ] Binary versions verified (no -dirty or -hash suffixes)
- [ ] Checksums saved from `dist/checksums.txt`
- [ ] GitHub release created with assets
- [ ] Homebrew formula manually edited with checksums from Step 4
- [ ] Formula checksums verified against GitHub release downloads
- [ ] Formula committed and pushed
- [ ] Homebrew installation tested successfully
- [ ] Installed version verified (`slack-butler --version`)

### Docker Release Checklist (Optional, after Go Release)

- [ ] Git tag verified to exist
- [ ] Docker image built locally (`make docker-build`)
- [ ] Local Docker test passed (`docker run astrostl/slack-butler:v1.X.X --version`)
- [ ] Logged in to Docker Hub (`docker login`)
- [ ] Multi-platform images pushed (`make docker-push`)
- [ ] Docker Hub publication verified (check tags page)
- [ ] Pull test successful (`docker pull astrostl/slack-butler:latest`)
- [ ] Version verification passed (`docker run astrostl/slack-butler:latest --version`)
- [ ] Multi-platform support confirmed (amd64 + arm64)

---

## Version Strategy

- **Beta releases**: `1.x.x-beta` for feature-complete testing
- **Stable releases**: `1.x.x` for production-ready versions
- **Major versions**: Breaking changes or significant feature additions

---

## Distribution Methods

After a successful release, users can install slack-butler via:

1. **Homebrew** (macOS, recommended):
   ```bash
   brew install astrostl/slack-butler/slack-butler
   ```

2. **Docker** (all platforms, containerized):
   ```bash
   docker pull astrostl/slack-butler:latest
   docker run astrostl/slack-butler:latest channels detect --token=$SLACK_TOKEN --since=7
   ```

3. **Go Install** (cross-platform):
   ```bash
   go install github.com/astrostl/slack-butler@latest
   ```

4. **Build from Source** (any platform):
   ```bash
   git clone https://github.com/astrostl/slack-butler.git
   cd slack-butler
   git checkout v1.X.X
   go build -o slack-butler
   ```

---

## Notes

- The entire release process should take about 10-20 minutes if everything goes smoothly
- Go module releases only require a git tag (very quick)
- Homebrew releases add GitHub release creation and formula updates
- Docker releases add image building and multi-platform publishing to Docker Hub
- Most issues come from committing changes after creating the release tag
- Always verify binary versions before creating the GitHub release (Homebrew only)
- Always verify Docker image versions before pushing to Docker Hub
- The Homebrew formula checksums MUST match the GitHub release assets exactly
- Never move the release tag after creating the GitHub release (Homebrew only)
- NEVER skip quality gates - every failure must be fixed before release
- Quality is non-negotiable - users depend on stable, secure code

---

## Release Strategy Summary

**Required for ALL releases:**
- Git tag (enables Go module distribution)
- Quality gates passed
- Documentation updated

**Optional distribution channels:**
- **Homebrew**: Requires GitHub release + formula update (macOS users)
- **Docker**: Requires Docker Hub push (containerized deployments, all platforms)

**Recommended approach:**
1. Start with Go module release (git tag) - enables `go install`
2. Add Homebrew release if targeting macOS users
3. Add Docker release if targeting container/cloud deployments

All three can be done for the same version tag.
