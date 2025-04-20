package ashby

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
)

var (
	// nolint:gochecknoglobals,lll
	supportPagination = datautils.NewSet("application", "applicationFeedback", "candidate",
		"candidateTag", "customField", "interview", "interviewSchedule", "interviewerPool", "job", "jobTemplate", "offer", "opening",
		"project", "surveyFormDefinition", "user")

	//nolint:gochecknoglobals
	supportSince = datautils.NewSet("application", "applicationFeedback", "interviewSchedule")
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	//nolint:lll
	writeSupport := []string{
		"application",
		"candidate",
		"candidateTag",
		"customField",
		"department",
		"interviewSchedule",
		"interviewerPool",
		"job",
		"location",
		"offer",
		"opening",
		"referral",
		"surveyRequest",
		"surveySubmission",
		"webhook",
	}

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
