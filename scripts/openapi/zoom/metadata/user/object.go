package user

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/internal/files"
)

var (
	allowedEndpoints = filters.NewRules( // nolint:gochecknoglobals
		"/groups",
		"/users",
		"/contacts/groups",
	)

	objectEndpoints = core.NewTrackedMap(map[string]string{ // nolint:gochecknoglobals
		"/contacts/groups": "contacts_groups",
	})
)

func Objects() pipeline.Pipeline[spec.Schema] {
	extractor, err := files.UsersFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithAutoSelectArrayItems(),
	)
	goutils.MustBeNil(err)

	pipe := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(
			filters.NewAllowPathStrategy(allowedEndpoints),
		)).
		Map(mapping.ObjectNameOverride(objectEndpoints)).
		Reduce(reducers.ShortestNameFromUrl).
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		))

	output.PrintUnusedRules("user/allowedEndpoints", allowedEndpoints)
	output.PrintRedundantKeys("user/objectEndpoints", objectEndpoints)

	return pipe
}
