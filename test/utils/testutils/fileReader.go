package testutils

import (
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

type FileData []byte

// DataFromFile is used for server mocking.
// Files must be located under ./test directory relative to the test runner.
func DataFromFile(t *testing.T, testFileName string) FileData {
	data, err := internalDataFromFile(testFileName)
	if err != nil {
		t.Fatalf("failed to start test, input file missing, %v", err)
	}

	return data
}

func internalDataFromFile(testFileName string) (FileData, error) {
	// NOTE: the deeper the call stack the higher the number should be
	_, parentCallerLocation, _, _ := runtime.Caller(2) // nolint:dogsled

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	relativePath, _ := strings.CutPrefix(parentCallerLocation, workingDir)
	testDir := path.Join(".", relativePath, "../test")

	return os.ReadFile(testDir + "/" + testFileName)
}
