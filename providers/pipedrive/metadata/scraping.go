package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	FileManager = scrapper.NewMetadataFileManager(schemas, fileconv.NewSiblingFileLocator()) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = FileManager.MustLoadSchemas() // nolint:gochecknoglobals
)
