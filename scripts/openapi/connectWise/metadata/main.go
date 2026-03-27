package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/connectWise/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals,lll
var (
	ignoreEndpoints = []string{
		// Schemas that have NO array properties
		"/company/addressFormats/count",
		"/company/companies/count",
		"/company/companies/statuses/count",
		"/company/companies/types/count",
		"/company/companyPickerItems/count",
		"/company/configurations/count",
		"/company/configurations/statuses/count",
		"/company/configurations/types/count",
		"/company/contacts/count",
		"/company/contacts/departments/count",
		"/company/contacts/relationships/count",
		"/company/contacts/types/count",
		"/company/countries/count",
		"/company/entityTypes/count",
		"/company/managedDevicesIntegrations/count",
		"/company/management/count",
		"/company/managementBackups/count",
		"/company/managementItSolutions/count",
		"/company/marketDescriptions/count",
		"/company/noteTypes/count",
		"/company/ownershipTypes/count",
		"/company/portalConfigurations/count",
		"/company/portalConfigurations/invoiceSetup/paymentProcessors/count",
		"/company/portalSecurityLevels/count",
		"/company/portalSecuritySettings/count",
		"/company/states/count",
		"/company/teamRoles/count",
		"/company/tracks/count",
	}
	displayNameOverride = map[string]string{
		// "external": "External connections",
	}
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	objects := Objects()
	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(files.OutputConnectWise.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputConnectWise.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	api3.PrintObjectsWithMultipleArrays()
	api3.PrintObjectsWithNoArrays()
	api3.PrintObjectsWithAutoSelectedArrays()

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputConnectWise.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			collectionDisplayName,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride, nil,
	)
	goutils.MustBeNil(err)

	return readObjects
}

func collectionDisplayName(displayName string) string {
	collectionName, found := strings.CutPrefix(displayName, "Collection Of ")
	if !found {
		return displayName
	}

	return naming.NewPluralString(collectionName).String()
}
