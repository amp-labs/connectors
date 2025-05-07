package instantlyai

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/instantlyai/metadata"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	writeSupport := []string{
		"accounts",
		"campaigns",
		"emails",
		"lead-lists",
		"inbox-placement-tests",
		"api-keys",
		"leads",
		"custom-tags",
		"block-lists-entries",
		"lead-labels",
		"workspace-group-members",
		"workspace-members",
		"subsequences",
		"email-verification"}

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
