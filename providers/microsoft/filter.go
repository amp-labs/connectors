package microsoft

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// List of objects to response fields.
// These fields are used for incremental reading.
//
// See `connectors/providers/microsoft/internal/metadata/main/main.go` script
// which helps with producing this registry.
// nolint:lll,gochecknoglobals
var incrementalObjects = map[string]string{
	"accessPackages":                         "modifiedDateTime",
	"acronyms":                               "lastModifiedDateTime",
	"admin/serviceAnnouncement/messages":     "lastModifiedDateTime",
	"alerts":                                 "lastModifiedDateTime",
	"alerts_v2":                              "lastUpdateDateTime",
	"androidManagedAppProtections":           "lastModifiedDateTime",
	"appRoleAssignments":                     "createdDateTime",
	"applications":                           "createdDateTime",
	"applications/microsoft.graph.delta()":   "createdDateTime",
	"articles":                               "lastUpdatedDateTime",
	"assignmentPolicies":                     "modifiedDateTime",
	"assignmentRequests":                     "createdDateTime",
	"assignmentScheduleRequests":             "createdDateTime",
	"assignmentSchedules":                    "modifiedDateTime",
	"authenticationStrengthPolicies":         "modifiedDateTime",
	"authorities":                            "createdDateTime",
	"bookingBusinesses":                      "lastUpdatedDateTime",
	"bookmarks":                              "lastModifiedDateTime",
	"bundles":                                "lastModifiedDateTime",
	"callRecords":                            "lastModifiedDateTime",
	"cases":                                  "lastModifiedDateTime",
	"catalogs":                               "modifiedDateTime",
	"categories":                             "createdDateTime",
	"certificateBasedAuthConfigurations":     "lastModifiedDateTime",
	"chats":                                  "lastUpdatedDateTime",
	"chats/microsoft.graph.getAllMessages()": "lastModifiedDateTime",
	"chats/microsoft.graph.getAllRetainedMessages()": "lastModifiedDateTime",
	"children":                      "lastModifiedDateTime",
	"citations":                     "createdDateTime",
	"communications/onlineMeetings": "creationDateTime",
	"communications/onlineMeetings/microsoft.graph.getAllRecordings(meetingOrganizerUserId='@meetingOrganizerUserId',startDateTime=@startDateTime,endDateTime=@endDateTime)":  "createdDateTime",
	"communications/onlineMeetings/microsoft.graph.getAllTranscripts(meetingOrganizerUserId='@meetingOrganizerUserId',startDateTime=@startDateTime,endDateTime=@endDateTime)": "createdDateTime",
	"conditionalAccessPolicies":                                        "modifiedDateTime",
	"connectedOrganizations":                                           "modifiedDateTime",
	"containerTypes":                                                   "createdDateTime",
	"containers":                                                       "createdDateTime",
	"customTaskExtensions":                                             "lastModifiedDateTime",
	"defaultManagedAppProtections":                                     "lastModifiedDateTime",
	"definitions":                                                      "lastModifiedDateTime",
	"delegatedAdminRelationships":                                      "lastModifiedDateTime",
	"deletedContainers":                                                "createdDateTime",
	"departments":                                                      "createdDateTime",
	"deviceAppManagement/managedAppRegistrations":                      "createdDateTime",
	"deviceCompliancePolicies":                                         "lastModifiedDateTime",
	"deviceConfigurations":                                             "lastModifiedDateTime",
	"deviceEnrollmentConfigurations":                                   "lastModifiedDateTime",
	"deviceImages":                                                     "lastModifiedDateTime",
	"deviceManagement/virtualEndpoint/cloudPCs":                        "lastModifiedDateTime",
	"directory/deletedItems/microsoft.graph.application":               "createdDateTime",
	"directory/deletedItems/microsoft.graph.group":                     "createdDateTime",
	"directory/deletedItems/microsoft.graph.user":                      "createdDateTime",
	"directory/subscriptions":                                          "createdDateTime",
	"documentSetVersions":                                              "lastModifiedDateTime",
	"drive/createdByUser/serviceProvisioningErrors":                    "createdDateTime",
	"drive/items":                                                      "lastModifiedDateTime",
	"drive/lastModifiedByUser/serviceProvisioningErrors":               "createdDateTime",
	"drive/list/createdByUser/serviceProvisioningErrors":               "createdDateTime",
	"drive/list/items":                                                 "lastModifiedDateTime",
	"drive/list/items/microsoft.graph.delta()":                         "lastModifiedDateTime",
	"drive/list/lastModifiedByUser/serviceProvisioningErrors":          "createdDateTime",
	"drive/list/operations":                                            "createdDateTime",
	"drive/microsoft.graph.recent()":                                   "lastModifiedDateTime",
	"drive/root/createdByUser/serviceProvisioningErrors":               "createdDateTime",
	"drive/root/lastModifiedByUser/serviceProvisioningErrors":          "createdDateTime",
	"drive/root/listItem/createdByUser/serviceProvisioningErrors":      "createdDateTime",
	"drive/root/listItem/lastModifiedByUser/serviceProvisioningErrors": "createdDateTime",
	"drive/root/listItem/versions":                                     "lastModifiedDateTime",
	"drive/root/microsoft.graph.delta()":                               "lastModifiedDateTime",
	"drive/root/versions":                                              "lastModifiedDateTime",
	"driveInclusionRules":                                              "lastModifiedDateTime",
	"driveProtectionUnits":                                             "lastModifiedDateTime",
	"driveProtectionUnitsBulkAdditionJobs":                             "lastModifiedDateTime",
	"drives":                                                           "lastModifiedDateTime",
	"ediscoveryCases":                                                  "lastModifiedDateTime",
	"education/me/assignments":                                         "lastModifiedDateTime",
	"education/me/assignments/microsoft.graph.delta()":                 "lastModifiedDateTime",
	"education/me/user/serviceProvisioningErrors":                      "createdDateTime",
	"eligibilityScheduleRequests":                                      "createdDateTime",
	"eligibilitySchedules":                                             "modifiedDateTime",
	"emailMethods":                                                     "createdDateTime",
	"endUserNotifications":                                             "lastModifiedDateTime",
	"engagementAsyncOperations":                                        "createdDateTime",
	"exchangeProtectionPolicies":                                       "lastModifiedDateTime",
	"exchangeRestoreSessions":                                          "lastModifiedDateTime",
	"externalAuthenticationMethods":                                    "createdDateTime",
	"fido2Methods":                                                     "createdDateTime",
	"filePlanReferences":                                               "createdDateTime",
	"followedSites":                                                    "lastModifiedDateTime",
	"following":                                                        "lastModifiedDateTime",
	"groups":                                                           "createdDateTime",
	"groups/microsoft.graph.delta()":                                   "createdDateTime",
	"healthIssues":                                                     "lastModifiedDateTime",
	"historyDefinitions":                                               "createdDateTime",
	"identity/conditionalAccess/authenticationStrength/policies":       "modifiedDateTime",
	"identity/conditionalAccess/policies":                              "modifiedDateTime",
	"identityGovernance/entitlementManagement/resources":               "modifiedDateTime",
	"identityGovernance/lifecycleWorkflows/deletedItems":               "lastModifiedDateTime",
	"identityGovernance/lifecycleWorkflows/deletedItems/workflows":     "lastModifiedDateTime",
	"identityGovernance/lifecycleWorkflows/workflows":                  "lastModifiedDateTime",
	"incidents":             "lastUpdateDateTime",
	"informationProtection": "createdDateTime",
	"internetExplorerMode":  "lastModifiedDateTime",
	"invitations/invitedUser/serviceProvisioningErrors": "createdDateTime",
	"iosManagedAppProtections":                          "lastModifiedDateTime",
	"issues":                                            "lastModifiedDateTime",
	"joinedTeams":                                       "createdDateTime",
	"landingPages":                                      "lastModifiedDateTime",
	"loginPages":                                        "lastModifiedDateTime",
	"mailboxInclusionRules":                             "lastModifiedDateTime",
	"mailboxProtectionUnits":                            "lastModifiedDateTime",
	"mailboxProtectionUnitsBulkAdditionJobs":            "lastModifiedDateTime",
	"managedAppPolicies":                                "lastModifiedDateTime",
	"managedEBooks":                                     "lastModifiedDateTime",
	"manifests":                                         "createdDateTime",
	"mdmWindowsInformationProtectionPolicies":           "lastModifiedDateTime",
	"me/activities":                                     "lastModifiedDateTime",
	"me/activities/microsoft.graph.recent()":            "lastModifiedDateTime",
	"me/authentication/operations":                      "createdDateTime",
	"me/calendar/calendarView":                          "lastModifiedDateTime",
	"me/calendar/calendarView/microsoft.graph.delta()":  "lastModifiedDateTime",
	"me/calendar/events":                                "lastModifiedDateTime",
	"me/calendar/events/microsoft.graph.delta()":        "lastModifiedDateTime",
	"me/calendarView":                                   "lastModifiedDateTime",
	"me/calendarView/microsoft.graph.delta()":           "lastModifiedDateTime",
	"me/chats":                                          "lastUpdatedDateTime",
	"me/chats/microsoft.graph.getAllMessages()":         "lastModifiedDateTime",
	"me/chats/microsoft.graph.getAllRetainedMessages()": "lastModifiedDateTime",
	"me/cloudClipboard/items":                           "lastModifiedDateTime",
	"me/cloudPCs":                                       "lastModifiedDateTime",
	"me/contacts":                                       "lastModifiedDateTime",
	"me/contacts/microsoft.graph.delta()":               "lastModifiedDateTime",
	"me/directReports/microsoft.graph.user":             "createdDateTime",
	"me/drives":                                         "lastModifiedDateTime",
	"me/events":                                         "lastModifiedDateTime",
	"me/events/microsoft.graph.delta()":                 "lastModifiedDateTime",
	"me/joinedTeams/microsoft.graph.getAllMessages()":   "lastModifiedDateTime",
	"me/managedAppRegistrations":                        "createdDateTime",
	"me/memberOf/microsoft.graph.group":                 "createdDateTime",
	"me/messages":                                       "lastModifiedDateTime",
	"me/messages/microsoft.graph.delta()":               "lastModifiedDateTime",
	"me/onenote/operations":                             "createdDateTime",
	"me/onlineMeetings":                                 "creationDateTime",
	"me/onlineMeetings/microsoft.graph.getAllRecordings(meetingOrganizerUserId='@meetingOrganizerUserId',startDateTime=@startDateTime,endDateTime=@endDateTime)":  "createdDateTime",
	"me/onlineMeetings/microsoft.graph.getAllTranscripts(meetingOrganizerUserId='@meetingOrganizerUserId',startDateTime=@startDateTime,endDateTime=@endDateTime)": "createdDateTime",
	"me/ownedDevices/microsoft.graph.appRoleAssignment":      "createdDateTime",
	"me/ownedObjects/microsoft.graph.application":            "createdDateTime",
	"me/ownedObjects/microsoft.graph.group":                  "createdDateTime",
	"me/planner/plans":                                       "createdDateTime",
	"me/planner/tasks":                                       "createdDateTime",
	"me/registeredDevices/microsoft.graph.appRoleAssignment": "createdDateTime",
	"me/serviceProvisioningErrors":                           "createdDateTime",
	"me/transitiveMemberOf/microsoft.graph.group":            "createdDateTime",
	"methods":                                                  "createdDateTime",
	"microsoft.graph.androidLobApp":                            "lastModifiedDateTime",
	"microsoft.graph.androidStoreApp":                          "lastModifiedDateTime",
	"microsoft.graph.driveProtectionUnit":                      "lastModifiedDateTime",
	"microsoft.graph.getAllEnterpriseInteractions()":           "createdDateTime",
	"microsoft.graph.getAllOnlineMeetingMessages()":            "lastModifiedDateTime",
	"microsoft.graph.getAllSites()":                            "lastModifiedDateTime",
	"microsoft.graph.getManagedAppPolicies()":                  "lastModifiedDateTime",
	"microsoft.graph.iosLobApp":                                "lastModifiedDateTime",
	"microsoft.graph.iosStoreApp":                              "lastModifiedDateTime",
	"microsoft.graph.iosVppApp":                                "lastModifiedDateTime",
	"microsoft.graph.macOSDmgApp":                              "lastModifiedDateTime",
	"microsoft.graph.macOSLobApp":                              "lastModifiedDateTime",
	"microsoft.graph.mailboxProtectionUnit":                    "lastModifiedDateTime",
	"microsoft.graph.managedAndroidLobApp":                     "lastModifiedDateTime",
	"microsoft.graph.managedIOSLobApp":                         "lastModifiedDateTime",
	"microsoft.graph.managedMobileLobApp":                      "lastModifiedDateTime",
	"microsoft.graph.microsoftStoreForBusinessApp":             "lastModifiedDateTime",
	"microsoft.graph.sharedWithMe()":                           "lastModifiedDateTime",
	"microsoft.graph.siteProtectionUnit":                       "lastModifiedDateTime",
	"microsoft.graph.win32LobApp":                              "lastModifiedDateTime",
	"microsoft.graph.windowsAppX":                              "lastModifiedDateTime",
	"microsoft.graph.windowsMobileMSI":                         "lastModifiedDateTime",
	"microsoft.graph.windowsUniversalAppX":                     "lastModifiedDateTime",
	"microsoft.graph.windowsWebApp":                            "lastModifiedDateTime",
	"microsoftAuthenticatorMethods":                            "createdDateTime",
	"mobileAppCategories":                                      "lastModifiedDateTime",
	"mobileAppConfigurations":                                  "lastModifiedDateTime",
	"mobileApps":                                               "lastModifiedDateTime",
	"namedLocations":                                           "modifiedDateTime",
	"notebooks":                                                "lastModifiedDateTime",
	"notificationMessageTemplates":                             "lastModifiedDateTime",
	"oneDriveForBusinessProtectionPolicies":                    "lastModifiedDateTime",
	"oneDriveForBusinessRestoreSessions":                       "lastModifiedDateTime",
	"organization":                                             "createdDateTime",
	"pages":                                                    "lastModifiedDateTime",
	"passwordMethods":                                          "createdDateTime",
	"payloads":                                                 "lastModifiedDateTime",
	"phoneMethods":                                             "createdDateTime",
	"planner/plans":                                            "createdDateTime",
	"planner/tasks":                                            "createdDateTime",
	"platformCredentialMethods":                                "createdDateTime",
	"print/operations":                                         "createdDateTime",
	"print/shares":                                             "createdDateTime",
	"privacy/subjectRightsRequests":                            "lastModifiedDateTime",
	"protectionPolicies":                                       "lastModifiedDateTime",
	"protectionUnits":                                          "lastModifiedDateTime",
	"qnas":                                                     "lastModifiedDateTime",
	"recoveryKeys":                                             "createdDateTime",
	"reflectCheckInResponses":                                  "createdDateTime",
	"reports/partners/billing/operations":                      "createdDateTime",
	"resourceEnvironments":                                     "modifiedDateTime",
	"resourceRequests":                                         "createdDateTime",
	"resourceRoleScopes":                                       "createdDateTime",
	"restoreSessions":                                          "lastModifiedDateTime",
	"retentionEventTypes":                                      "lastModifiedDateTime",
	"retentionEvents":                                          "lastModifiedDateTime",
	"retentionLabels":                                          "lastModifiedDateTime",
	"riskDetections":                                           "lastUpdatedDateTime",
	"roleManagement/directory/roleAssignmentScheduleRequests":  "createdDateTime",
	"roleManagement/directory/roleAssignmentSchedules":         "modifiedDateTime",
	"roleManagement/directory/roleEligibilityScheduleRequests": "createdDateTime",
	"roleManagement/directory/roleEligibilitySchedules":        "modifiedDateTime",
	"roleManagement/entitlementManagement/roleAssignmentScheduleRequests":  "createdDateTime",
	"roleManagement/entitlementManagement/roleAssignmentSchedules":         "modifiedDateTime",
	"roleManagement/entitlementManagement/roleEligibilityScheduleRequests": "createdDateTime",
	"roleManagement/entitlementManagement/roleEligibilitySchedules":        "modifiedDateTime",
	"roleManagementPolicies":                 "lastModifiedDateTime",
	"rubrics":                                "lastModifiedDateTime",
	"sectionGroups":                          "lastModifiedDateTime",
	"sections":                               "lastModifiedDateTime",
	"secureScoreControlProfiles":             "lastModifiedDateTime",
	"secureScores":                           "createdDateTime",
	"security/attackSimulation/operations":   "createdDateTime",
	"security/subjectRightsRequests":         "lastModifiedDateTime",
	"sensors":                                "createdDateTime",
	"serviceApps":                            "lastModifiedDateTime",
	"servicePrincipalRiskDetections":         "lastUpdatedDateTime",
	"sharePointProtectionPolicies":           "lastModifiedDateTime",
	"sharePointRestoreSessions":              "lastModifiedDateTime",
	"shares":                                 "lastModifiedDateTime",
	"signIns":                                "createdDateTime",
	"simulationAutomations":                  "lastModifiedDateTime",
	"simulations":                            "lastModifiedDateTime",
	"siteInclusionRules":                     "lastModifiedDateTime",
	"siteLists":                              "lastModifiedDateTime",
	"siteProtectionUnits":                    "lastModifiedDateTime",
	"siteProtectionUnitsBulkAdditionJobs":    "lastModifiedDateTime",
	"sites":                                  "lastModifiedDateTime",
	"sites/microsoft.graph.delta()":          "lastModifiedDateTime",
	"softwareOathMethods":                    "createdDateTime",
	"special":                                "lastModifiedDateTime",
	"targetedManagedAppConfigurations":       "lastModifiedDateTime",
	"teams":                                  "createdDateTime",
	"teams/microsoft.graph.getAllMessages()": "lastModifiedDateTime",
	"teamwork/deletedTeams/microsoft.graph.getAllMessages()": "lastModifiedDateTime",
	"temporaryAccessPassMethods":                             "createdDateTime",
	"termsAndConditions":                                     "lastModifiedDateTime",
	"threatAssessmentRequests":                               "createdDateTime",
	"trainings":                                              "lastModifiedDateTime",
	"trending":                                               "lastModifiedDateTime",
	"triggerTypes":                                           "lastModifiedDateTime",
	"triggers":                                               "lastModifiedDateTime",
	"userConfigurations":                                     "modifiedDateTime",
	"userExperienceAnalyticsBaselines":                       "createdDateTime",
	"userRegistrationDetails":                                "lastUpdatedDateTime",
	"userSettings":                                           "lastModifiedDateTime",
	"users":                                                  "createdDateTime",
	"users/microsoft.graph.delta()":                          "createdDateTime",
	"vppTokens":                                              "lastModifiedDateTime",
	"vulnerabilities":                                        "lastModifiedDateTime",
	"whoisHistoryRecords":                                    "lastUpdateDateTime",
	"whoisRecords":                                           "lastUpdateDateTime",
	"windowsHelloForBusinessMethods":                         "createdDateTime",
	"windowsInformationProtectionPolicies":                   "lastModifiedDateTime",
	"workforceIntegrations":                                  "lastModifiedDateTime",
}

// https://learn.microsoft.com/en-us/graph/filter-query-parameter?tabs=http
type filterQuery struct {
	since string
	until string
}

func (q filterQuery) Since(objectName string, timestamp time.Time) filterQuery {
	if timestamp.IsZero() {
		return q
	}

	fieldName, ok := incrementalObjects[objectName]
	if !ok {
		// Object does not support incremental reading.
		return q
	}

	value := datautils.Time.FormatRFC3339inUTCWithMilliseconds(timestamp)
	q.since = fmt.Sprintf("%v ge %v", fieldName, value)

	return q
}

func (q filterQuery) Until(objectName string, timestamp time.Time) filterQuery {
	if timestamp.IsZero() {
		return q
	}

	fieldName, ok := incrementalObjects[objectName]
	if !ok {
		// Object does not support incremental reading.
		return q
	}

	value := datautils.Time.FormatRFC3339inUTCWithMilliseconds(timestamp)
	q.until = fmt.Sprintf("%v le %v", fieldName, value)

	return q
}

func (q filterQuery) String() string {
	if q.since == "" && q.until == "" {
		return ""
	}

	if q.since != "" && q.until != "" {
		return fmt.Sprintf("%v and %v", q.since, q.until)
	}

	if q.since != "" {
		return q.since
	}

	return q.until
}
