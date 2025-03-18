package staticschema

import (
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv"
)

const SchemasFile = "schemas.json"

// FileManager operates on schema.json file.
type FileManager[F FieldMetadataMap, C any] struct {
	schemas []byte
	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewFileManager[F FieldMetadataMap, C any](schemas []byte, locator fileconv.FileLocator) *FileManager[F, C] {
	return &FileManager[F, C]{
		schemas: schemas,
		locator: locator,
	}
}

// SaveSchemas is useful method when creating or updating static schema file.
func (m FileManager[F, C]) SaveSchemas(schemas *Metadata[F, C]) error {
	schemas.refactorLongestCommonPath()

	return m.FlushSchemas(schemas)
}

// FlushSchemas stores schemas into the file without any modifications.
func (m FileManager[F, C]) FlushSchemas(schemas *Metadata[F, C]) error {
	return m.flush.ToFile(m.locator.AbsPathTo(SchemasFile), schemas)
}

// MustLoadSchemas parses static schema file and returns Metadata for deep connector.
func (m FileManager[F, C]) MustLoadSchemas() *Metadata[F, C] {
	var result *Metadata[F, C]

	err := json.Unmarshal(m.schemas, &result)
	if err != nil {
		// This error should never occur if schemas file is of correct format.
		// If at least one test exists for the connector this will be caught at development time.
		return &Metadata[F, C]{}
	}

	return result
}
