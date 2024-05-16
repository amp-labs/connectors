package metadata

import (
	"path/filepath"
	"runtime"

	"github.com/amp-labs/connectors/common/scrapper"
)

const (
	IndexFile   = "index.json"
	SchemasFile = "schemas.json"
)

func SaveIndex(index *scrapper.ModelURLRegistry) error {
	return scrapper.FlushToFile(resolveRelativePath(IndexFile), index)
}

func LoadIndex() (*scrapper.ModelURLRegistry, error) {
	var registry *scrapper.ModelURLRegistry

	err := scrapper.LoadFile(resolveRelativePath(IndexFile), &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func SaveSchemas(schemas *scrapper.ObjectMetadataResult) error {
	return scrapper.FlushToFile(resolveRelativePath(SchemasFile), schemas)
}

func LoadSchemas() (*scrapper.ObjectMetadataResult, error) {
	var schemas *scrapper.ObjectMetadataResult

	err := scrapper.LoadFile(resolveRelativePath(SchemasFile), &schemas)
	if err != nil {
		return nil, err
	}

	return schemas, nil
}

func resolveRelativePath(filename string) string {
	_, thisMethodsLocation, _, _ := runtime.Caller(0) // nolint:dogsled
	localDir := filepath.Dir(thisMethodsLocation)

	return localDir + "/" + filename
}
