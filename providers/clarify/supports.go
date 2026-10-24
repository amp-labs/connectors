package clarify

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	objectNameCompany = "company"
	objectNamePerson  = "person"
	objectNameDeal    = "deal"
	objectNameMeeting = "meeting"
	objectNameTask    = "task"
	// Workspace members, GET /v1/workspaces/{slug}/users.
	objectNameUsers = "users"

	// Custom objects are named with a `c_` prefix ("Sales Order" becomes `c_sales_order`) and use
	// the same endpoints as built-in objects.
	// https://developer.clarify.ai/docs/guides/schemas-and-relationships#custom-objects
	customObjectPrefix = "c_"
)

// Built-in objects, read from GET /v1/workspaces/{slug}/objects/{object}/resources.
// https://developer.clarify.ai/docs/getting-started/key-concepts#objects
var builtInObjects = []string{ // nolint:gochecknoglobals
	objectNameCompany,
	objectNamePerson,
	objectNameDeal,
	objectNameMeeting,
	objectNameTask,
}

// liveTestCustomObject is the custom object in Ampersand's Clarify test workspace. It is listed so the live
// tests read one `c_*` object; any workspace's custom objects are read through the `c_*` registry entry.
const liveTestCustomObject = "c_sample_contract"

// SupportedReadObjects lists the objects the live tests read: the built-in objects, users, and one custom
// object. Custom objects (`c_*`) are read in general, but which ones exist depends on the workspace.
func SupportedReadObjects() []string {
	return append(slices.Clone(builtInObjects), objectNameUsers, liveTestCustomObject)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(SupportedReadObjects(), ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: customObjectPrefix + "*",
				Support:  components.ReadSupport,
			},
		},
	}
}

// isRecordObject reports whether the object is read from /objects/{object}/resources.
func isRecordObject(objectName string) bool {
	return slices.Contains(builtInObjects, objectName) || strings.HasPrefix(objectName, customObjectPrefix)
}

func isReadableObject(objectName string) bool {
	return isRecordObject(objectName) || objectName == objectNameUsers
}
