package scrapper

import (
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv"
)

const (
	IndexFile           = "index.json"
	SchemasFile         = "schemas.json"
	QueryParamStatsFile = "queryParamStats.json"
)

type MetadataFileManager struct {
	schemas []byte
	locator fileconv.FileLocator
}

func NewMetadataFileManager(schemas []byte, locator fileconv.FileLocator) *MetadataFileManager {
	return &MetadataFileManager{
		schemas: schemas,
		locator: locator,
	}
}

func (m MetadataFileManager) SaveIndex(index *ModelURLRegistry) error {
	index.Sort()

	return FlushToFile(m.locator.AbsPathTo(IndexFile), index)
}

func (m MetadataFileManager) LoadIndex() (*ModelURLRegistry, error) {
	var registry *ModelURLRegistry

	err := LoadFile(m.locator.AbsPathTo(IndexFile), &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (m MetadataFileManager) SaveSchemas(schemas *ObjectMetadataResult) error {
	return FlushToFile(m.locator.AbsPathTo(SchemasFile), schemas)
}

func (m MetadataFileManager) LoadSchemas() (*ObjectMetadataResult, error) {
	var result *ObjectMetadataResult

	err := json.Unmarshal(m.schemas, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m MetadataFileManager) SaveQueryParamStats(stats *QueryParamStats) error {
	return FlushToFile(m.locator.AbsPathTo(QueryParamStatsFile), stats)
}
