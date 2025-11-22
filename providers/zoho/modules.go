package zoho

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

var supportedModules = common.Modules{ //nolint: gochecknoglobals
	common.ModuleRoot: common.Module{
		ID:    common.ModuleRoot,
		Label: "",
	},
	providers.ModuleZohoCRM: common.Module{
		ID:    providers.ModuleZohoCRM,
		Label: "zoho CRM",
	},
	providers.ModuleZohoServiceDeskPlus: common.Module{
		ID:    providers.ModuleZohoServiceDeskPlus,
		Label: "zoho serviceDesk Plus",
	},
	providers.ModuleZohoDesk: common.Module{
		ID:    providers.ModuleZohoDesk,
		Label: "zoho Desk",
	},
}
