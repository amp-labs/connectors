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
	apiFile []byte

	InputConnectWise  = api2.NewOpenapiFileManager[any](apiFile)
	OutputConnectWise = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/connectWise/internal/metadata"))
)
