package user

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var allowedEndpoints = []string{ // nolint:gochecknoglobals
	"/groups",
	"/users",
}

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
		nil, nil, api3.IdenticalObjectLocator,
	)

	goutils.MustBeNil(err)

	return objects
}
