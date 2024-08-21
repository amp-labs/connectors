package fileconv

import (
	"path/filepath"
	"runtime"
)

// FileLocator provides an absolute path to the file.
// This is useful for writing data to that file.
// Loading data from file should be done via go embed.
type FileLocator interface {
	AbsPathTo(filename string) string
}

type SiblingFileLocator struct {
	directory string
}

// NewSiblingFileLocator creates locator, which is capable of resolving full path to files
// located under the same directory where this constructor was called.
func NewSiblingFileLocator() *SiblingFileLocator {
	_, thisMethodsLocation, _, _ := runtime.Caller(1) // nolint:dogsled
	directory := filepath.Dir(thisMethodsLocation)

	return &SiblingFileLocator{
		directory: directory,
	}
}

func (l SiblingFileLocator) AbsPathTo(filename string) string {
	return filepath.Join(l.directory, filename)
}
