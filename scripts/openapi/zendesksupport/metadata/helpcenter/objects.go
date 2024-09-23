package helpcenter

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zendesksupport/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"/api/v2/guide/search", // -> this is unified search for Article, Post, ExternalRecord
		"/api/v2/help_center/sessions",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/api/v2/help_center/articles/search":        "articles",
		"/api/v2/help_center/articles/labels":        "article_labels",
		"/api/v2/help_center/community_posts/search": "community_posts",
	}
)

func Objects() []api3.Schema {
	explorer, err := openapi.HelpCenterFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	handy.Must(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		api3.CustomMappingObjectCheck(zendesksupport.ObjectNameToResponseField[zendesksupport.ModuleHelpCenter]),
	)
	handy.Must(err)

	return objects
}
