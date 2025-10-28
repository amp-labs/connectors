package scrapper

import (
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
)

const (
	IndexFile           = "index.json"
	QueryParamStatsFile = "queryParamStats.json"
)

type MetadataFileManager[F staticschema.FieldMetadataMap, C any] struct {
	staticschema.FileManager[F, C]

	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewExtendedMetadataFileManager[F staticschema.FieldMetadataMap, C any](
	schemas []byte, locator fileconv.FileLocator,
) *MetadataFileManager[F, C] {
	return &MetadataFileManager[F, C]{
		FileManager: *staticschema.NewFileManager[F, C](schemas, locator),
		locator:     locator,
	}
}

// NewMetadataFileManager allows reading and writing schema.json.
// Use dedicated instances of NewReader or NewWriter.
// Deprecated.
func NewMetadataFileManager[F staticschema.FieldMetadataMap](
	schemas []byte, locator fileconv.FileLocator,
) *MetadataFileManager[F, any] {
	return NewExtendedMetadataFileManager[F, any](schemas, locator)
}

func NewReader[F staticschema.FieldMetadataMap](schemas []byte) *MetadataFileManager[F, any] {
	return NewExtendedMetadataFileManager[F, any](schemas, nil)
}

func NewWriter[F staticschema.FieldMetadataMap](locator fileconv.FileLocator) *MetadataFileManager[F, any] {
	return NewExtendedMetadataFileManager[F, any](nil, locator)
}

func (m MetadataFileManager[F, C]) SaveIndex(index *ModelURLRegistry) error {
	index.Sort()

	return m.flush.ToFile(m.locator.AbsPathTo(IndexFile), index)
}

func (m MetadataFileManager[F, C]) LoadIndex() (*ModelURLRegistry, error) {
	var registry *ModelURLRegistry

	err := LoadFile(m.locator.AbsPathTo(IndexFile), &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (m MetadataFileManager[F, C]) SaveQueryParamStats(stats *QueryParamStats) error {
	return m.flush.ToFile(m.locator.AbsPathTo(QueryParamStatsFile), stats)
}
