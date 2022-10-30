package utils

import "fmt"

var Version string = "0"
var Commit string = "unknown"

func VersionString() string {
	return fmt.Sprintf("v%s-%s", Version, Commit[:7])
}
