package mapping

import (
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func RemoveURLPrefix(prefix string) pipeline.MapFunc[spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		schema.URLPath, _ = strings.CutPrefix(schema.URLPath, prefix)

		return schema
	}
}
