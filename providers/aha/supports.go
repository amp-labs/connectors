package aha

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/aha/metadata"
)

const (
	objectNameAudits            = "audits"
	objectNameHistoricalAudits  = "historical_audits"
	objectNameIdeasEndorsements = "ideas/endorsements"
	objectNameIdeas             = "ideas"
)

var supportSince = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAudits,
	objectNameHistoricalAudits,
	objectNameIdeasEndorsements,
	objectNameIdeas,
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
