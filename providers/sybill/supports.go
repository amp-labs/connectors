package sybill

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectNameConversations = "conversations"
	objectNameAccounts      = "accounts"
	objectNameDeals         = "deals"
)

var supportedObjects = []string{ // nolint:gochecknoglobals
	objectNameConversations,
	objectNameAccounts,
	objectNameDeals,
}

// Each object names its identifier after itself.
// https://api.sybill.ai/docs/data-models.html
var idFields = datautils.Map[string, string]{ // nolint:gochecknoglobals
	objectNameConversations: "conversationId",
	objectNameAccounts:      "accountId",
	objectNameDeals:         "dealId",
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
