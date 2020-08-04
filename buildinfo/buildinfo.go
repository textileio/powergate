package buildinfo

import "fmt"

var (
	// GitCommit is set by govvv at build time.
	GitCommit = "default-git-commit"
	// GitBranch  is set by govvv at build time.
	GitBranch = "default-git-branch"
	// GitState  is set by govvv at build time.
	GitState = "default-git-state"
	// GitSummary is set by govvv at build time.
	GitSummary = "default-git-summary"
	// BuildDate  is set by govvv at build time.
	BuildDate = "default-build-date"
	// Version  is set by govvv at build time.
	Version = "default-version"
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
