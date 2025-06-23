package main

import (
	"github.com/astrostl/slack-buddy-ai/cmd"
)

// Version information injected at build time via GoReleaser.
var (
	Version   = "dev"     // Version is the application version
	BuildTime = "unknown" // BuildTime is when the binary was built
	GitCommit = "unknown" // GitCommit is the git commit hash
)

func main() {
	cmd.Execute(Version, BuildTime, GitCommit)
}
