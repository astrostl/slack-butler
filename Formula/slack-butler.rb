class SlackButler < Formula
  desc "CLI tool to help Slack workspaces be more useful and tidy"
  homepage "https://github.com/astrostl/slack-butler"
  version "1.5.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.5.0/slack-butler-v1.5.0-darwin-arm64.tar.gz"
    sha256 "4f927d1f4491e6ef12d9c43baf457bc2618bb00af0f3f764974f33e9251b7909"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/slack-butler/releases/download/v1.5.0/slack-butler-v1.5.0-darwin-amd64.tar.gz"
    sha256 "e2642fb6e103989b41907e7889ac6f5810cb999e23ebcea389a528a11ef2c821"
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
