package files

import (
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var OutputSalesforcePardot = scrapper.NewWriter[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
	fileconv.NewPath("providers/salesforce/internal/pardot"))
