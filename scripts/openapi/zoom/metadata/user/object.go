package user

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	allowedEndpoints = []string{ // nolint:gochecknoglobals
		"/groups",
		"/users",
		"/contacts/groups",
	}

	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"contacts/groups": "Contact Groups",
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
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowedEndpoints),
		objectEndpoints, displayNameOverride, nil,
	)

	goutils.MustBeNil(err)

	return objects
}
