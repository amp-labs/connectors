package metadata

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
	"runtime"

	"github.com/amp-labs/connectors/common/scrapper"
)

const (
	IndexFile   = "index.json"
	SchemasFile = "schemas.json"
)

// Static file containing a list of object metadata is embedded and can be served.
//
//go:embed schemas.json
var schemas []byte

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
	var result *scrapper.ObjectMetadataResult

	err := json.Unmarshal(schemas, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func resolveRelativePath(filename string) string {
	_, thisMethodsLocation, _, _ := runtime.Caller(0) // nolint:dogsled
	localDir := filepath.Dir(thisMethodsLocation)

	return localDir + "/" + filename
}
