package scrapper

import (
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
)

const (
	IndexFile           = "index.json"
	QueryParamStatsFile = "queryParamStats.json"
)

type MetadataFileManager[F staticschema.FieldMetadataMap] struct {
	staticschema.FileManager[F]

	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewMetadataFileManager[F staticschema.FieldMetadataMap](
	schemas []byte, locator fileconv.FileLocator,
) *MetadataFileManager[F] {
	return &MetadataFileManager[F]{
		FileManager: *staticschema.NewFileManager[F](schemas, locator),
		locator:     locator,
	}
}

func (m MetadataFileManager[F]) SaveIndex(index *ModelURLRegistry) error {
	index.Sort()

	return m.flush.ToFile(m.locator.AbsPathTo(IndexFile), index)
}

func (m MetadataFileManager[F]) LoadIndex() (*ModelURLRegistry, error) {
	var registry *ModelURLRegistry

	err := LoadFile(m.locator.AbsPathTo(IndexFile), &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (m MetadataFileManager[F]) SaveQueryParamStats(stats *QueryParamStats) error {
	return m.flush.ToFile(m.locator.AbsPathTo(QueryParamStatsFile), stats)
}
