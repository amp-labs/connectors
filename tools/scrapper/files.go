package scrapper

import "encoding/json"

const (
	IndexFile           = "index.json"
	SchemasFile         = "schemas.json"
	QueryParamStatsFile = "queryParamStats.json"
)

// MetadataFileLocator locates index and schema files.
// Every module stores these files in its own place.
type MetadataFileLocator interface {
	AbsPathTo(filename string) string
}

type MetadataFileManager struct {
	schemas []byte
	locator MetadataFileLocator
}

func NewMetadataFileManager(schemas []byte, locator MetadataFileLocator) *MetadataFileManager {
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
