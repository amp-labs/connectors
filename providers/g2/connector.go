package g2

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	//go:embed schema.json
	schemaContent []byte

	fileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2](
		schemaContent, fileconv.NewSiblingFileLocator())

	schemas = fileManager.MustLoadSchemas()
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider

	// subjectProductId represents either subject_product_id or product_id.
	subjectProductId string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.G2, params, constructor)
	if err != nil {
		return nil, err
	}

	connector.subjectProductId = params.Metadata["subject_product_id"]

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"subject_product_id"},
		},
	}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), schemas)

	return connector, nil
}
