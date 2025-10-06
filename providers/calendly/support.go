package calendly

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	requiresOrgURIQueryParam = datautils.NewSet("activity_log_entries", //nolint: gochecknoglobals
		"event_types", "groups", "organization_memberships", "outgoing_communications",
		"routing_forms", "scheduled_events", "group_relationships")

	requiresUserURIQueryParam = datautils.NewSet("user_busy_times", //nolint: gochecknoglobals
		"user_availability_schedules",
		"locations", "scheduled_events")

	EndpointWithUpdatedAtParam = datautils.NewSet("event_types") //nolint: gochecknoglobals
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}

	writeSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
