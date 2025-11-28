package output

import (
	"log/slog"
	"path/filepath"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/tools/scrapper"
)

const QueryParamStatsFile = "queryParamStats.json"

func WriteQueryParamStats(dirName string, pipe pipeline.Pipeline[spec.Schema]) error {
	if err := validate(pipe); err != nil {
		return err
	}

	return Write(
		filepath.Join(dirName, QueryParamStatsFile),
		extractQueryParamStats(pipe),
	)
}

func extractQueryParamStats(pipe pipeline.Pipeline[spec.Schema]) any {
	registry := datautils.NamedLists[string]{}

	for _, object := range pipe.List() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"path", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		for _, queryParam := range object.QueryParams {
			objectID := object.Operation + " " + object.URLPath
			registry.Add(queryParam, objectID)
		}
	}

	return scrapper.CalculateQueryParamStats(registry)
}
