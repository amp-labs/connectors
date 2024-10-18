package scrapper

import (
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
)

const (
	IndexFile           = "index.json"
	QueryParamStatsFile = "queryParamStats.json"
)

type MetadataFileManager struct {
	staticschema.FileManager

	locator fileconv.FileLocator
	flush   fileconv.Flusher
}

func NewMetadataFileManager(schemas []byte, locator fileconv.FileLocator) *MetadataFileManager {
	return &MetadataFileManager{
		FileManager: *staticschema.NewFileManager(schemas, locator),
		locator:     locator,
	}
}

func (m MetadataFileManager) SaveIndex(index *ModelURLRegistry) error {
	index.Sort()

	return m.flush.ToFile(m.locator.AbsPathTo(IndexFile), index)
}

func (m MetadataFileManager) LoadIndex() (*ModelURLRegistry, error) {
	var registry *ModelURLRegistry

	err := LoadFile(m.locator.AbsPathTo(IndexFile), &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (m MetadataFileManager) SaveQueryParamStats(stats *QueryParamStats) error {
	return m.flush.ToFile(m.locator.AbsPathTo(QueryParamStatsFile), stats)
}
