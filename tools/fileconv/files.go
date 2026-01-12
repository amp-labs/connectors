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

type CallLevel int

const (
	LevelSibling CallLevel = iota
	LevelParent
)

type LevelFileLocator struct {
	directory string
}

// NewSiblingFileLocator creates locator, which is capable of resolving full path to files
// located under the same directory where this constructor was called.
func NewSiblingFileLocator() *LevelFileLocator {
	// One is added to count this method call.
	// Sibling is not relative to this method but to the caller.
	return NewLevelFileLocator(LevelSibling + 1)
}

// NewLevelFileLocator creates locator, which is capable of resolving full path to files
// located relative to the call stack.
func NewLevelFileLocator(level CallLevel) *LevelFileLocator {
	_, thisMethodsLocation, _, _ := runtime.Caller(1 + int(level)) // nolint:dogsled
	directory := filepath.Dir(thisMethodsLocation)

	return &LevelFileLocator{
		directory: directory,
	}
}

func (l LevelFileLocator) AbsPathTo(filename string) string {
	return filepath.Join(l.directory, filename)
}

type FilePath struct {
	path string
}

func NewPath(path string) *FilePath {
	return &FilePath{path: path}
}

func (f FilePath) AbsPathTo(filename string) string {
	return filepath.Join(f.path, filename)
}
