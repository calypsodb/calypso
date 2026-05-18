package internal

import (
	"fmt"
)

var (
	// Version release version
	Version = "0.1.0"

	// Commit will be overwritten automatically by the build system
	Commit = "HEAD"
)

// FullVersion returns the full version and commit hash
func FullVersion() string {
	return fmt.Sprintf("%s@%s", Version, Commit)
}
