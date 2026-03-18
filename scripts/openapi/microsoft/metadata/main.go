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
	"github.com/amp-labs/connectors/scripts/openapi/microsoft/internal/files"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals,lll
var (
	ignoreEndpoints = []string{
		// Schemas -- MULTIPLE arrays
		"/admin/serviceAnnouncement",
		"/admin/sharepoint/settings",
		"/auditLogs",
		"/communications",
		"/deviceAppManagement",
		"/deviceManagement",
		"/deviceManagement/conditionalAccessSettings",
		"/deviceManagement/userExperienceAnalyticsAppHealthOverview",
		"/deviceManagement/virtualEndpoint",
		"/directory",
		"/drive",
		"/drive/createdByUser",
		"/drive/lastModifiedByUser",
		"/drive/list",
		"/drive/list/createdByUser",
		"/drive/list/drive",
		"/drive/list/lastModifiedByUser",
		"/drive/root",
		"/drive/root/createdByUser",
		"/drive/root/lastModifiedByUser",
		"/drive/root/listItem",
		"/drive/root/listItem/createdByUser",
		"/drive/root/listItem/driveItem",
		"/drive/root/listItem/lastModifiedByUser",
		"/education",
		"/education/me",
		"/education/me/user",
		"/education/reports",
		"/employeeExperience",
		"/identity",
		"/identity/conditionalAccess/authenticationStrength",
		"/identity/riskPrevention",
		"/identityGovernance/accessReviews",
		"/identityGovernance/entitlementManagement",
		"/identityGovernance/lifecycleWorkflows",
		"/identityGovernance/privilegedAccess/group",
		"/identityGovernance/termsOfUse",
		"/identityProtection",
		"/invitations/invitedUser",
		"/me",
		"/me/authentication",
		"/me/drive",
		"/me/employeeExperience",
		"/me/insights",
		"/me/onenote",
		"/me/planner",
		"/me/settings/workHoursAndLocations",
		"/me/teamwork",
		"/planner",
		"/policies",
		"/policies/crossTenantAccessPolicy",
		"/print",
		"/reports",
		"/reports/partners/billing",
		"/roleManagement/directory",
		"/roleManagement/entitlementManagement",
		"/search",
		"/security",
		"/security/attackSimulation",
		"/security/identities",
		"/security/labels",
		"/security/threatIntelligence",
		"/solutions",
		"/solutions/backupRestore",
		"/solutions/virtualEvents",
		"/storage/fileStorage",
		"/teamwork",
		"/tenantRelationships",
		// Schemas that have NO array properties
		"/admin",
		"/admin/edge",
		"/admin/exchange",
		"/admin/microsoft365Apps",
		"/admin/microsoft365Apps/installationOptions",
		"/admin/people/itemInsights",
		"/admin/people/pronouns",
		"/admin/reportSettings",
		"/admin/sharepoint",
		"/compliance",
		"/copilot/admin",
		"/copilot/admin/settings",
		"/copilot/admin/settings/limitedMode",
		"/copilot/interactionHistory",
		"/copilot/reports",
		"/deviceAppManagement/managedAppRegistrations/microsoft.graph.getUserIdsWithFlaggedAppRegistration()",
		"/deviceManagement/applePushNotificationCertificate",
		"/deviceManagement/applePushNotificationCertificate/microsoft.graph.downloadApplePushNotificationCertificateSigningRequest()",
		"/deviceManagement/auditEvents/microsoft.graph.getAuditCategories()",
		"/deviceManagement/deviceCompliancePolicyDeviceStateSummary",
		"/deviceManagement/deviceConfigurationDeviceStateSummaries",
		"/deviceManagement/managedDeviceOverview",
		"/deviceManagement/microsoft.graph.userExperienceAnalyticsSummarizeWorkFromAnywhereDevices()",
		"/deviceManagement/softwareUpdateStatusSummary",
		"/deviceManagement/userExperienceAnalyticsWorkFromAnywhereHardwareReadinessMetric",
		"/deviceManagement/virtualEndpoint/auditEvents/microsoft.graph.getAuditActivityTypes()",
		"/deviceManagement/virtualEndpoint/report",
		"/directory/federationConfigurations/microsoft.graph.availableProviderTypes()",
		"/drive/createdByUser/mailboxSettings",
		"/drive/lastModifiedByUser/mailboxSettings",
		"/drive/list/createdByUser/mailboxSettings",
		"/drive/list/lastModifiedByUser/mailboxSettings",
		"/drive/root/createdByUser/mailboxSettings",
		"/drive/root/lastModifiedByUser/mailboxSettings",
		"/drive/root/listItem/createdByUser/mailboxSettings",
		"/drive/root/listItem/fields",
		"/drive/root/listItem/lastModifiedByUser/mailboxSettings",
		"/drive/root/retentionLabel",
		"/education/me/user/mailboxSettings",
		"/identity/identityProviders/microsoft.graph.availableProviderTypes()",
		"/identityGovernance",
		"/identityGovernance/entitlementManagement/settings",
		"/identityGovernance/lifecycleWorkflows/insights",
		"/identityGovernance/lifecycleWorkflows/settings",
		"/identityGovernance/privilegedAccess",
		"/identityProviders/microsoft.graph.availableProviderTypes()",
		"/invitations/invitedUser/mailboxSettings",
		"/me/dataSecurityAndGovernance/protectionScopes",
		"/me/licenseDetails/microsoft.graph.getTeamsLicensingDetails()",
		"/me/mailboxSettings",
		"/me/manager",
		"/me/manager/$ref",
		"/me/microsoft.graph.exportDeviceAndAppManagementData()",
		"/me/microsoft.graph.getManagedDevicesWithAppFailures()",
		"/me/onPremisesSyncBehavior",
		"/me/photo",
		"/me/presence",
		"/me/settings/itemInsights",
		"/me/settings/storage",
		"/me/solutions",
		"/me/solutions/workingTimeSchedule",
		"/policies/authenticationFlowsPolicy",
		"/policies/authorizationPolicy",
		"/policies/crossTenantAccessPolicy/default",
		"/policies/crossTenantAccessPolicy/templates",
		"/policies/crossTenantAccessPolicy/templates/multiTenantOrganizationIdentitySynchronization",
		"/policies/crossTenantAccessPolicy/templates/multiTenantOrganizationPartnerConfiguration",
		"/policies/defaultAppManagementPolicy",
		"/policies/deviceRegistrationPolicy",
		"/policies/identitySecurityDefaultsEnforcementPolicy",
		"/reports/microsoft.graph.deviceConfigurationDeviceActivity()",
		"/reports/microsoft.graph.deviceConfigurationUserActivity()",
		"/reports/microsoft.graph.managedDeviceEnrollmentFailureDetails()",
		"/reports/microsoft.graph.managedDeviceEnrollmentTopFailures()",
		"/reports/partners",
		"/reports/partners/billing/reconciliation",
		"/reports/partners/billing/reconciliation/billed",
		"/reports/partners/billing/reconciliation/unbilled",
		"/reports/partners/billing/usage",
		"/reports/partners/billing/usage/billed",
		"/reports/partners/billing/usage/unbilled",
		"/reports/security",
		"/roleManagement",
		"/security/dataSecurityAndGovernance/protectionScopes",
		"/security/identities/sensorCandidateActivationConfiguration",
		"/security/identities/sensors/microsoft.graph.security.getDeploymentAccessKey()",
		"/security/identities/sensors/microsoft.graph.security.getDeploymentPackageUri()",
		"/storage",
		"/storage/settings",
		"/teamwork/teamsAppSettings",
		"/tenantRelationships/multiTenantOrganization/joinRequest",
		//
		// Excluded endpoints. If this was done by mistake, each endpoint can be examined one more time.
		// These endpoints were added by following a general trend of not being relevant as a collection of objects.
		//
		"/admin/people",
		"/admin/teams",
		"/admin/teams/policy",
		"/authenticationMethodsPolicy",
		"/deviceManagement/reports",
		"/directory/publicKeyInfrastructure",
		"/drive/root/analytics",
		"/drive/root/analytics/allTime",
		"/drive/root/analytics/lastSevenDays",
		"/drive/root/listItem/analytics",
		"/identityGovernance/appConsent",
		"/informationProtection/bitlocker",
		"/me/calendar",
		"/me/cloudClipboard",
		"/me/inferenceClassification",
		"/me/outlook",
		"/me/outlook/microsoft.graph.supportedLanguages()",
		"/me/outlook/microsoft.graph.supportedTimeZones()",
		"/me/settings",
		"/me/settings/shiftPreferences",
		"/me/settings/storage/quota",
		"/policies/adminConsentRequestPolicy",
		"/policies/authenticationMethodsPolicy",
		"/privacy",
		"/reports/authenticationMethods",
		"/reports/authenticationMethods/microsoft.graph.usersRegisteredByFeature()",
		"/reports/authenticationMethods/microsoft.graph.usersRegisteredByMethod()",
		"/security/dataSecurityAndGovernance",
		"/storage/settings/quota",
	}
	displayNameOverride = map[string]string{
		"external":                                                    "External connections",
		"informationProtection":                                       "Threat Assessment Requests",
		"internetExplorerMode":                                        "Internet Explorer Sites",
		"me/dataSecurityAndGovernance":                                "Data Security And Governance Sensitivity Labels",
		"me/dataSecurityAndGovernance/activities":                     "Data Security And Governance Activities",
		"microsoft.graph.getAttackSimulationRepeatOffenders()":        "Attack Simulation Repeat Offenders",
		"microsoft.graph.getAttackSimulationSimulationUserCoverage()": "Attack Simulation User Coverage",
		"microsoft.graph.getManagedAppDiagnosticStatuses()":           "Managed App Diagnostic Statuses",
		"microsoft.graph.getSourceImages()":                           "Source Images",
		"multiTenantOrganization":                                     "Tenants",
		"todo":                                                        "Todo Lists",
		"tracing":                                                     "Message Tracing",
		"triggers":                                                    "Retention Events",
		"userExperienceAnalyticsOverview":                             "User Experience Analytics Insights",
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

	goutils.MustBeNil(files.OutputMicrosoftGraph.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputMicrosoftGraph.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	api3.PrintObjectsWithMultipleArrays()
	api3.PrintObjectsWithNoArrays()
	// api3.PrintObjectsWithAutoSelectedArrays()

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputMicrosoftGraph.GetExplorer(
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
