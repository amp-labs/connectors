package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api2"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed esignature-api.json
	openapiAPI []byte

	InputDocusignESignature  = api2.NewOpenapiFileManager[any](openapiAPI)
	OutputDocusignESignature = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/docusign/metadata"))
)
