package utils

import "fmt"

var version string = "0"
var commit string = "unknown"

func VersionString() string {
	return fmt.Sprintf("v%s-%s", version, commit[:7])
}
