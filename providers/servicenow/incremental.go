package servicenow

import (
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// snDateTimeFormat is ServiceNow's encoded-query datetime layout.
const snDateTimeFormat = "2006-01-02 15:04:05"

// incrementalScopedObjects are non-Table-API objects whose list GET accepts a
// sysparm_query on sys_updated_on, so they support Since/Until delta reads.
var incrementalScopedObjects = []string{
	"account",
	"case",
	"change",
	"consumer",
	"contact",
	"lead",
}

// supportsIncrementalRead reports whether the object can be filtered by
// sys_updated_on via sysparm_query. Every Table API object qualifies, as do a few
// scoped APIs that accept sysparm_query. TMF/Open (bare-array) and SCIM responses
// don't expose sysparm_query, so for those a read with Since simply returns the
// full set.
func supportsIncrementalRead(objectName string) bool {
	if slices.Contains(incrementalScopedObjects, objectName) {
		return true
	}

	return strings.HasPrefix(objectPaths[objectName], "now/table/")
}

// incrementalQuery builds the sys_updated_on sysparm_query fragment for the
// [Since, Until] window, or returns "" when no incremental filter applies.
//
// Note: ServiceNow interprets the datetime in the integration user's timezone, so
// that user's timezone should be UTC for the window to line up exactly.
func incrementalQuery(params common.ReadParams) string {
	if params.Since.IsZero() || !supportsIncrementalRead(params.ObjectName) {
		return ""
	}

	updates := []string{"sys_updated_on>=" + params.Since.UTC().Format(snDateTimeFormat)}

	if !params.Until.IsZero() {
		updates = append(updates, "sys_updated_on<="+params.Until.UTC().Format(snDateTimeFormat))
	}

	return strings.Join(updates, "^")
}
