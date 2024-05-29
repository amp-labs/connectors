package metadata

import (
	_ "embed"
	"path/filepath"
	"runtime"

	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager(schemas, locator{}) // nolint:gochecknoglobals
)

type locator struct{}

func (locator) AbsPathTo(filename string) string {
	_, thisMethodsLocation, _, _ := runtime.Caller(0) // nolint:dogsled
	localDir := filepath.Dir(thisMethodsLocation)

	return localDir + "/" + filename
}
