package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed spec3.yaml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals

	OutputStripe = scrapper.NewWriter[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		fileconv.NewPath("providers/stripe/internal/metadata"))
)
