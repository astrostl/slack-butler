class SlackButler < Formula
  desc "CLI tool to help Slack workspaces be more useful and tidy"
  homepage "https://github.com/astrostl/slack-butler"
  version "1.3.1"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.3.1/slack-butler-v1.3.1-darwin-arm64.tar.gz"
    sha256 "0cd17ff18e2e7e0182e89020118ff3e8cdbf8e1be7dada54a251b8d819df2174"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.3.1/slack-butler-v1.3.1-darwin-amd64.tar.gz"
    sha256 "139bb3754334b7398b7e024d49ee05c0264deabb8fc5d760adc1d9d21e38ecd0"
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
