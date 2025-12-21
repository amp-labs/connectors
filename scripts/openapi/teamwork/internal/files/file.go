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
	//go:embed projects-api-v3.oas2.yml
	openAPI []byte

	InputTeamwork  = api2.NewOpenapiFileManager[any](openAPI)
	OutputTeamwork = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/teamwork/internal/metadata"))
)
