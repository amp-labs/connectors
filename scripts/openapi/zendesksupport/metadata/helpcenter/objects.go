package helpcenter

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
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
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"articles":        "results",
		"article_labels":  "labels",
		"community_posts": "results",
	}, func(objectName string) (fieldName string) {
		return objectName
	})
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.HelpCenterFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	return objects
}
