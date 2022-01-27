package version

import "time"

var (
	Version       = "1.0.0"
	GitCommit     = "HEAD"                              // git rev-parse [--short] HEAD
	GitBranch     = "master"                            // git symbolic-ref -q --short HEAD
	GitSummary    = "HEAD"                              // git describe --tags --dirty --always
	GitCommitTime = ""                                  // git log -1 --format=%cd --date=format:'%a %b %d %Y %H:%M:%S GMT%z'
	BuildTime     = time.Now().Format(time.RFC3339Nano) // date +"%a %b %d %Y %H:%M:%S GMT%z"
)
