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
	//go:embed api.json
	apiFile []byte

	InputPipedrive  = api3.NewOpenapiFileManager[any](apiFile)
	OutputPipedrive = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/pipeliner/internal/metadata"))
)
