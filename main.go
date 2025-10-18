package main

import (
	"runtime/debug"

	"github.com/astrostl/slack-butler/cmd"
)

// Version information injected at build time via ldflags.
var (
	Version   = "dev"     // Version is the application version
	BuildTime = "unknown" // BuildTime is when the binary was built
	GitCommit = "unknown" // GitCommit is the git commit hash
)

func main() {
	// If version is still "dev", try to get it from build info (for go install)
	version := Version
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	cmd.Execute(version, BuildTime, GitCommit)
}
