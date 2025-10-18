class SlackButler < Formula
  desc "CLI tool to help Slack workspaces be more useful and tidy"
  homepage "https://github.com/astrostl/slack-butler"
  version "1.2.1"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.2.1/slack-butler-v1.2.1-darwin-arm64.tar.gz"
    sha256 "1f0f0adaa0c297d1abff7393eea7fd4aea89603da576640c2e570c2a1e622954"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.2.1/slack-butler-v1.2.1-darwin-amd64.tar.gz"
    sha256 "57653ab2d816f5fd612fe23df10ddddafa81d7f38be1662be345b94f9cd14100"
  else
    odie "slack-butler is only supported on macOS via Homebrew. Use 'go install' or build from source for other platforms."
  end

  def install
    bin.install "slack-butler-darwin-arm64" => "slack-butler" if Hardware::CPU.arm?
    bin.install "slack-butler-darwin-amd64" => "slack-butler" if Hardware::CPU.intel?
  end

  def caveats
    <<~EOS
      slack-butler requires a Slack Bot Token to function.

      Configuration:
        Create a .env file or export SLACK_TOKEN:
          export SLACK_TOKEN=xoxb-your-token-here

      Basic usage:
        slack-butler health                     # Verify configuration and connectivity
        slack-butler channels detect            # Detect new channels (last 8 days)
        slack-butler channels archive           # Archive inactive channels (dry-run)
        slack-butler channels highlight         # Highlight random active channels

      For detailed usage and required Slack permissions:
        https://github.com/astrostl/slack-butler

      Note: Most commands default to dry-run mode. Use --commit to apply changes.
    EOS
  end

  test do
    system bin/"slack-butler", "--help"
  end
end
