package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager(schemas, fileconv.NewSiblingFileLocator()) // nolint:gochecknoglobals
)
