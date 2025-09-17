package zoho

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

var supportedModules = common.Modules{
	providers.ZohoCRM: common.Module{
		ID:    providers.ZohoCRM,
		Label: "zoho CRM",
	},
	providers.ZohoDeskV2: common.Module{
		ID:    providers.ZohoDeskV2,
		Label: "zoho Desk",
	},
}
