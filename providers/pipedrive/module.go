package pipedrive

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

var supportedModules = common.Modules{ //nolint: gochecknoglobals
	common.ModuleRoot: common.Module{
		ID:    common.ModuleRoot,
		Label: "",
	},
	providers.PipedriveV1: common.Module{
		ID:    providers.PipedriveV1,
		Label: "Pipedrive V1",
	},
	providers.PipedriveV2: common.Module{
		ID:    providers.PipedriveV2,
		Label: "Pipedrive V2",
	},
}
