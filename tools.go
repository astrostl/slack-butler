//go:build tools

package tools

// This file tracks development tool dependencies.
// Tools are imported as blank imports to ensure they're included in go.mod
// but not compiled into the main binary due to the build constraint.

import (
	_ "github.com/fzipp/gocyclo/cmd/gocyclo"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
