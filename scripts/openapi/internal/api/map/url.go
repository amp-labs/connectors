package mapping

import (
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func RemoveUrlPrefix(prefix string) pipeline.MapFunc[spec.Schema, spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		schema.UrlPath, _ = strings.CutPrefix(schema.UrlPath, prefix)

		return schema
	}
}
