package people

import (
	"net/http"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/google/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// Requires query parameter `resourceNames=[]`.
		"/v1/contactGroups:batchGet",
		"/v1/people:batchGet",
		// Requires query parameter `query`, that is a text search.
		"/v1/otherContacts:search",
		"/v1/people:searchContacts",
		"/v1/people:searchDirectoryPeople",
		// URL that require IDs.
		// We are explicit, because we need some such endpoints.
		// The IDs will be hard coded making them regular endpoints.
		"/v1/{resourceName}",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/v1/people:listDirectoryPeople": "peopleDirectory",
	}
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.PeopleFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(
		http.MethodGet,
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		func(objectName, fieldName string) bool {
			return false
		},
	)
	goutils.MustBeNil(err)

	for index, object := range objects {
		if object.URLPath == "/v1/{resourceName}/connections" {
			// Override some values.
			object.URLPath = "/v1/people/me/connections"
			object.ObjectName = "myConnections"
			objects[index] = object
		}
	}

	return objects
}
