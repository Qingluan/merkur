package util

import "fmt"

// version info
var (
	Name        = "@github.com/Qingluan/merkur/github.com/Qingluan/merkur"
	Version     = "unknown"
	BuildTime   = "unknown"
	GitRevision = "unknown"
)

// PrintVersion print version
func PrintVersion() {
	fmt.Printf("name: %s\n", Name)
	fmt.Printf("version: %s\n", Version)
	fmt.Printf("buildTime: %s\n", BuildTime)
	fmt.Printf("gitRevision: %s\n", GitRevision)
}
