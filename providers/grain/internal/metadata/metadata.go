package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/objectcatalog"
)

var (
	// objectsMetadata is created manually, see README.md.
	//
	//go:embed objectsMetadata.json
	objectsMetadata []byte

	Catalog = objectcatalog.MustLoad(objectsMetadata) // nolint:gochecknoglobals
)
