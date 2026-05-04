package filters

import (
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func KeepWithPath(matcher PathMatcher) pipeline.FilterFunc[spec.Schema] {
	return func(schema spec.Schema) bool {
		if matcher == nil {
			return false
		}

		return matcher.IsPathMatching(schema.URLPath)
	}
}
