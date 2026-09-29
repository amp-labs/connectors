package bamboohr

import (
	"net/http"

	"github.com/amp-labs/connectors/internal/datautils"
)

// writeSupportedObjects lists objects that support create and/or update via the deep connector.
//
//nolint:gochecknoglobals
var writeSupportedObjects = datautils.NewStringSet(
	objectTimeOffRequests,
	objectTimeTrackingProjects,
	objectWebhooks,
)

// deleteSupportedObjects lists objects with a DELETE (or equivalent remove) API on the collection item path.
//
//nolint:gochecknoglobals
var deleteSupportedObjects = datautils.NewStringSet(
	objectTimeTrackingProjects,
	objectWebhooks,
)

// writeUpdateMethod maps object name to the HTTP method used when RecordId is set.
//
//nolint:gochecknoglobals
var writeUpdateMethod = datautils.NewDefaultMap(map[string]string{
	objectWebhooks:             http.MethodPut,
	objectTimeTrackingProjects: http.MethodPatch,
	objectTimeOffRequests:      http.MethodPatch,
}, func(string) string {
	return http.MethodPut
})

// writeUpdateContentType sets Content-Type for update requests when non-empty.
// time_off_requests expects JSON Merge Patch per OpenAPI.
//
//nolint:gochecknoglobals
var writeUpdateContentType = datautils.NewDefaultMap(map[string]string{
	objectTimeOffRequests: "application/merge-patch+json",
}, func(string) string {
	return "application/json"
})
