package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

const (
	ModuleCalendar common.ModuleID = "calendar"
	ModulePeople   common.ModuleID = "people"
)

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals

var moduleSubdomains = datautils.NewDefaultMap[common.ModuleID, string]( // nolint: gochecknoglobals
	map[common.ModuleID]string{
		ModulePeople: "people",
	},
	func(moduleID common.ModuleID) (subdomain string) {
		return "www"
	},
)

func getSubdomain(moduleID common.ModuleID) catalogreplacer.CustomCatalogVariable {
	subdomain := moduleSubdomains.Get(moduleID)

	return catalogreplacer.CustomCatalogVariable{
		Plan: catalogreplacer.SubstitutionPlan{
			From: "subdomain",
			To:   subdomain,
		},
	}
}
