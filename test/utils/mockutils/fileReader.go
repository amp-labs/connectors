package mockutils

import (
	"os"
	"path"
	"runtime"
	"strings"
)

// DataFromFile is used for server mocking.
// Files must be located under ./test directory relative to the test runner.
func DataFromFile(testFileName string) ([]byte, error) {
	_, parentCallerLocation, _, _ := runtime.Caller(1) // nolint:dogsled

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	relativePath, _ := strings.CutPrefix(parentCallerLocation, workingDir)
	testDir := path.Join(".", relativePath, "../test")

	return os.ReadFile(testDir + "/" + testFileName)
}
