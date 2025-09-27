package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api2"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing openapi spec.
	//go:embed openapi.json
	openapiAPI []byte

	InputSalesflare  = api2.NewOpenapiFileManager[any](openapiAPI)
	OutputSalesflare = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/salesflare/internal/metadata"))
)
