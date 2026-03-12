package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var OutputSalesforcePardot = scrapper.NewWriter[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
	fileconv.NewPath("providers/salesforce/internal/pardot"))

var (
	//go:embed exports-fields.json
	ExportsFields []byte
	//go:embed imports-fields.json
	ImportsFields []byte
)
