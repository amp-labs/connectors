package files

import (
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	OutputSalesloft = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/salesloft/internal/metadata"))
)
