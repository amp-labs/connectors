package googledicsovery

import (
	"encoding/json"
)

// FileManager locates google discovery file.
// Allows to read data of interest.
type FileManager struct {
	file []byte
}

func NewFileManager(file []byte) *FileManager {
	return &FileManager{
		file: file,
	}
}

func (m FileManager) GetExplorer(opts ...Option) (*Explorer, error) {
	discoverFile := Document{}

	if err := json.Unmarshal(m.file, &discoverFile); err != nil {
		return nil, err
	}

	return &Explorer{
		document:   discoverFile,
		parameters: createParams(opts),
	}, nil
}
