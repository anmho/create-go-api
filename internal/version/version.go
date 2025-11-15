package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the version of the application
	Version = "dev"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
)

// Info returns version information
func Info() string {
	return fmt.Sprintf("create-go-api version %s (commit: %s, built: %s, go: %s)", Version, Commit, Date, runtime.Version())
}

// Short returns short version string
func Short() string {
	return Version
}

