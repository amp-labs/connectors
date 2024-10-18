package staticschema

import (
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv"
)

const SchemasFile = "schemas.json"

// FileManager operates on schema.json file.
type FileManager struct {
	schemas []byte
	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewFileManager(schemas []byte, locator fileconv.FileLocator) *FileManager {
	return &FileManager{
		schemas: schemas,
		locator: locator,
	}
}

// SaveSchemas is useful method when creating or updating static schema file.
func (m FileManager) SaveSchemas(schemas *Metadata) error {
	schemas.refactorLongestCommonPath()

	return m.flush.ToFile(m.locator.AbsPathTo(SchemasFile), schemas)
}

// MustLoadSchemas parses static schema file and returns Metadata for deep connector.
func (m FileManager) MustLoadSchemas() *Metadata {
	var result *Metadata

	err := json.Unmarshal(m.schemas, &result)
	if err != nil {
		// This error should never occur if schemas file is of correct format.
		// If at least one test exists for the connector this will be caught at development time.
		return &Metadata{}
	}

	return result
}
