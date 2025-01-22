package staticschema

import (
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv"
)

const SchemasFile = "schemas.json"

// FileManager operates on schema.json file.
type FileManager[F FieldMetadataMap] struct {
	schemas []byte
	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewFileManager[F FieldMetadataMap](schemas []byte, locator fileconv.FileLocator) *FileManager[F] {
	return &FileManager[F]{
		schemas: schemas,
		locator: locator,
	}
}

// SaveSchemas is useful method when creating or updating static schema file.
func (m FileManager[F]) SaveSchemas(schemas *Metadata[F]) error {
	schemas.refactorLongestCommonPath()

	return m.flush.ToFile(m.locator.AbsPathTo(SchemasFile), schemas)
}

// MustLoadSchemas parses static schema file and returns Metadata for deep connector.
func (m FileManager[F]) MustLoadSchemas() *Metadata[F] {
	var result *Metadata[F]

	err := json.Unmarshal(m.schemas, &result)
	if err != nil {
		// This error should never occur if schemas file is of correct format.
		// If at least one test exists for the connector this will be caught at development time.
		return &Metadata[F]{}
	}

	return result
}
