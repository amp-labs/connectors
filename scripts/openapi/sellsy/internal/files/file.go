package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing openapi spec.
	//go:embed sellsy.v2.latest.yaml
	apiFile []byte

	InputSellsy  = api3.NewOpenapiFileManager[any](apiFile)
	OutputSellsy = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/sellsy/internal/metadata"))

	OpenAPIFile     = api.NewFile(apiFile)
	OutputSellsyDir = "providers/sellsy/internal/metadata"
)
