package cmd

import (
	"fmt"
)

var (
	// BuildCommit lastest build commit (set by Makefile)
	BuildCommit = ""
	// BuildTag if the `BuildCommit` matches a tag
	BuildTag = ""
	// BuildTime set by build script (set by Makefile)
	BuildTime = ""
)

func version() string {
	if BuildTag != "" {
		return fmt.Sprintf("%s/%s", BuildTag, BuildTime)
	}
	return fmt.Sprintf("%s/%s", BuildCommit, BuildTime)
}
