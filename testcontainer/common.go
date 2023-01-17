package testcontainer

import (
	"fmt"
	"path/filepath"
)

// DoubledPort - helper func for prepare port mapping.
func DoubledPort(port string) string {
	return fmt.Sprintf("%s:%[1]s", port) // output: "<port>:<port>"
}

// GetAbsPath - get absolute path based on pwd and relative path.
func GetAbsPath(relativePath string) string {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		fmt.Printf("could not resolve abs path for: %s. err is: %v", relativePath, err)
	}

	return path
}
