package circleback

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	objectNameMeetings    = "meetings"
	objectNameActionItems = "action-items"
	objectNameTag         = "tag"
	objectNamePeople      = "people"
	objectNameCompanies   = "companies"
)

var supportedObjects = []string{ // nolint:gochecknoglobals
	objectNameMeetings,
	objectNameActionItems,
	objectNameTag,
	objectNamePeople,
	objectNameCompanies,
}

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
