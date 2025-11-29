package output

import (
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func validate(pipe pipeline.Pipeline[spec.Schema]) error {
	uniqueNames := datautils.IndexedLists[string, string]{}

	for _, object := range pipe.List() {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		uniqueNames.Add(object.ObjectName, object.URLPath)
	}

	success := true

	for objectName, urls := range uniqueNames {
		if len(urls) > 1 {
			success = false

			slog.Error("object name is shared by more than one object",
				"objectName", objectName,
				"urls", urls,
			)
		}
	}

	if !success {
		return errors.New("objects have failed validation before writing to file") // nolint:err113
	}

	return nil
}
