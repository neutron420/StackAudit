package version

import "fmt"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func FullVersion() string {
	return fmt.Sprintf("StackAudit %s (commit %s, built %s)", Version, Commit, Date)
}
