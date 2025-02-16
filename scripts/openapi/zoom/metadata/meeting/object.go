package meeting

import (
	"log"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	allowedEndpoints = []string{ // nolint:gochecknoglobals
		// "/archive_files",
		"/devices/groups",
		"/devices",
		// "/meetings/meeting_summaries",
		// "/report/billing",
		// "/report/activities",
		"/sip_phones/phones",
		"/tracking_fields",
	}
)

func Objects() []metadatadef.Schema {
	log.Println("Starting user objects")
	explorer, err := openapi.MeetingFileManager.GetExplorer(
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
