package ashby

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
)

var (
	//nolint:gochecknoglobals,lll
	supportPagination = datautils.NewSet("application.list", "candidate.list", "interview.list", "feedbackFormDefinition.list",
		"job.list", "jobPosting.list", "offer.list")

	supportSince = datautils.NewSet("application.list") //nolint:gochecknoglobals
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
