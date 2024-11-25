package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static files containing openapi spec.
	//
	// Multiple files are downloaded from https://api.aweber.com/.
	//
	// Files preprocessing:
	// Since each module has a reference to external files, the main task was to consolidate everything
	// into self-sustained files. This was achieved using openapi-yaml generator in Intelij. Campaigns module
	// had a malformed yaml file which had to be fixed before feeding it into generator.
	// This method resolved majority of references, except to code samples file. This was removed manually via regex.
	//
	// Summary:
	// Each yaml file describes API for that module without any external references.

	//go:embed sections/accounts.yaml
	accountsApiFile []byte
	//go:embed sections/broadcasts.yaml
	broadcastsApiFile []byte
	//go:embed sections/campaigns.yaml
	campaignsApiFile []byte
	//go:embed sections/custom_fields.yaml
	customFieldsApiFile []byte
	//go:embed sections/integrations.yaml
	integrationsApiFile []byte
	//go:embed sections/landing_pages.yaml
	landingPagesApiFile []byte
	//go:embed sections/lists.yaml
	listsApiFile []byte
	//go:embed sections/segments.yaml
	segmentsApiFile []byte
	//go:embed sections/subscribers.yaml
	subscribersApiFile []byte
	//go:embed sections/webforms.yaml
	webformsApiFile []byte

	FileManagers = []*api3.OpenapiFileManager{ // nolint:gochecknoglobals
		api3.NewOpenapiFileManager(accountsApiFile),
		api3.NewOpenapiFileManager(broadcastsApiFile),
		api3.NewOpenapiFileManager(campaignsApiFile),
		api3.NewOpenapiFileManager(customFieldsApiFile),
		api3.NewOpenapiFileManager(integrationsApiFile),
		api3.NewOpenapiFileManager(landingPagesApiFile),
		api3.NewOpenapiFileManager(listsApiFile),
		api3.NewOpenapiFileManager(segmentsApiFile),
		api3.NewOpenapiFileManager(subscribersApiFile),
		api3.NewOpenapiFileManager(webformsApiFile),
	}
)
