package notion

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectUsers   = "users"
	objectUploads = "file_uploads"
)

// readEndpoint holds read implementation details that are not part of objectsMetadata.json.
// Field schemas and list URL paths live in the object catalog; responseKey and similar
// endpoint facts come from the OpenAPI script's endpoints.json (gitignored maintainer output).
// Keep this map in sync when go run ./scripts/openapi/notion/metadata changes endpoints.json.
type readEndpoint struct {
	// responseKey is the JSON property that holds the record array in a list response.
	responseKey string
	// timeField enables connector-side Since/Until filtering when the API has no time query param.
	// Empty means incremental time filters are not applied for this object.
	timeField string
}

//nolint:gochecknoglobals
var readRegistry = datautils.Map[string, readEndpoint]{
	objectUsers: {
		responseKey: "results",
	},
	objectUploads: {
		responseKey: "results",
		timeField:   uploadTimeField,
	},
}

//nolint:gochecknoglobals
var readSupportedObjects = readRegistry.KeySet()

var supportedReadObjects = readRegistry.Keys() //nolint:gochecknoglobals

// SupportedReadObjects lists object names implemented by Read.
func SupportedReadObjects() []string {
	return slices.Clone(supportedReadObjects)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportedReadObjects, ",")),
				Support:  components.ReadSupport,
			},
			{
				// POST /v1/file_uploads create only; users have no write API.
				Endpoint: objectUploads,
				Support:  components.WriteSupport,
			},
		},
	}
}

func readEndpointFor(objectName string) (readEndpoint, error) {
	spec, ok := readRegistry[objectName]
	if !ok {
		return readEndpoint{}, common.ErrOperationNotSupportedForObject
	}

	return spec, nil
}
