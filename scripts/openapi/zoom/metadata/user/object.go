package user

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	allowedEndpoints = []string{ // nolint:gochecknoglobals
		"/groups",
		"/users",
		"/contacts/groups",
	}

	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/contacts/groups": "contacts_groups",
	}
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.UsersFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowedEndpoints),
		objectEndpoints, nil, api3.CustomMappingObjectCheck(
			zoom.ObjectNameToResponseField[common.ModuleID(providers.ModuleZoomUser)],
		),
	)

	goutils.MustBeNil(err)

	return objects
}
