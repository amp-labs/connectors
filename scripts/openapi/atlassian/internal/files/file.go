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
	// This can be downloaded here:
	// https://developer.atlassian.com/cloud/confluence/rest/v2/intro/#auth
	// (search for "OpenAPI", it is not easy to spot).
	//go:embed rest-v2.json
	confluenceAPI []byte

	InputConfluenceV2  = api3.NewOpenapiFileManager[any](confluenceAPI)
	OutputConfluenceV2 = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/atlassian/internal/confluence"))
)
