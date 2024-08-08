package fileconv

// FileLocator provides an absolute path to the file.
// This is useful for writing data to that file.
// Loading data from file should be done via go embed.
type FileLocator interface {
	AbsPathTo(filename string) string
}
