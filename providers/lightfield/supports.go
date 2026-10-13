package lightfield

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectNameAccounts      = "accounts"
	objectNameContacts      = "contacts"
	objectNameOpportunities = "opportunities"
	// Workspace members (GET /v1/members); the customer's "users".
	objectNameMembers = "members"
)

var supportedObjects = []string{ // nolint:gochecknoglobals
	objectNameAccounts,
	objectNameContacts,
	objectNameOpportunities,
	objectNameMembers,
}

// Objects that expose GET /v1/{object}/definitions.
// https://docs.lightfield.app/using-the-api/fields-and-relationships#supported-object-types
var objectsWithDefinitions = datautils.NewSet( // nolint:gochecknoglobals
	objectNameAccounts,
	objectNameContacts,
	objectNameOpportunities,
)

// Objects whose list endpoint filters on `$updatedAt[greaterThan|lessThan]`.
// /v1/members documents only limit and offset and rejects filters (400 unrecognized_keys).
// https://docs.lightfield.app/using-the-api/list-endpoints#filtering
var objectsWithUpdatedAtFilter = datautils.NewSet( // nolint:gochecknoglobals
	objectNameAccounts,
	objectNameContacts,
	objectNameOpportunities,
)

// SupportedReadObjects lists the objects this connector reads.
func SupportedReadObjects() []string {
	return slices.Clone(supportedObjects)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportedObjects, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
