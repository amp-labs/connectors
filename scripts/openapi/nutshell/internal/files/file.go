package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing openapi spec.
	//go:embed openapi.json
	openapiAPI []byte

	InputNutshell  = api3.NewOpenapiFileManager[any](openapiAPI)
	OutputNutshell = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/nutshell/internal/metadata"))
)
