package metadata

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/objectcatalog"
)

var (
	//go:embed objectsMetadata.json
	objectsMetadata []byte
	Catalog         = objectcatalog.MustLoad(objectsMetadata) // nolint:gochecknoglobals
)
