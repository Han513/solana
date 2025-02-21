package utils

import "fmt"

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func GetVersionInfo() string {
	return fmt.Sprintf(
		"Version: %s\nBuild Time: %s\nGit Commit: %s",
		Version, BuildTime, GitCommit,
	)
}
