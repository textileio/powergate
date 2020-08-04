package buildinfo

import "fmt"

var (
	// GitCommit is set by govvv at build time.
	GitCommit = "n/a"
	// GitBranch  is set by govvv at build time.
	GitBranch = "n/a"
	// GitState  is set by govvv at build time.
	GitState = "n/a"
	// GitSummary is set by govvv at build time.
	GitSummary = "n/a"
	// BuildDate  is set by govvv at build time.
	BuildDate = "n/a"
	// Version  is set by govvv at build time.
	Version = "n/a"
)

// Summary prints a summary of all build info.
func Summary() string {
	return fmt.Sprintf(
		"\tversion:\t%s\n\tbuild date:\t%s\n\tgit summary:\t%s\n\tgit branch:\t%s\n\tgit commit:\t%s\n\tgit state:\t%s",
		Version,
		BuildDate,
		GitSummary,
		GitBranch,
		GitCommit,
		GitState,
	)
}
