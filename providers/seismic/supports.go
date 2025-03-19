package seismic

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		// These are supported objects in the reporting API,
		// ref: https://developer.seismic.com/seismicsoftware/reference/h1-reporting-api-overview
		"adminImpersonationSessions", "contentUsageHistory", "contentViewHistory", "dailyActiveUsers",
		"distributionApprovalWorkflows", "distributionApprovalWorkflowStepsHistory", "interactionContexts",
		"interactionRecipients", "notificationFrequencySettings", "notificationStatus", "userActivity",
		"virusScanAuditLog", "aiActivity", "aiGeneratedText", "aiGeneratedTextUserFeedback", "aiSuggestedContentProperties",
		"copilotForSalesDsrRecommendations", "copilotForSalesRecommendations", "announcements", "channels", "posts",
		"postContents", "contents", "contentActivities", "contentAskExperts", "contentInsertInstances", "contentPages",
		"contentReviews", "contentSlideInsertInstances", "contentVersions", "customProperties", "customPropertyAssignments",
		"meetings", "meetingAgendaUpdates", "meetingContentPagePresentations", "meetingContentPresentations",
		"meetingGeneralNotesUpdates", "meetingKeywords", "meetingParticipants", "meetingQuestions", "meetingTrackers",
		"pageClicks", "pageContentHistory", "watermarks", "emails", "emailTemplateInstances",
		"emailTemplateSectionSelections", "emailTemplateStaticImages", "emailTemplateVariableValues", "answers",
		"assignments", "customUserFields", "customUserFieldValues", "groupManagers", "feedbackCriteria",
		"instructorLedTrainingEvents", "instructorLedTrainingEventAttendance", "instructorLedTrainingEventContentAssignments",
		"trainingGroupManagers", "instructorLedTrainingEventSessions", "learningJourneys", "learningJourneySteps",
		"learningJourneyTasks", "learningProgress", "learningStatuses", "lessons",
		"lessonTags", "lessonVersions", "paths", "pathContents", "pathTags", "proficiencyLevels",
		"questions", "skills", "skillAssessments", "skillProfiles", "skillProfileDetails", "skillRatings",
		"skillRequests", "skillReviews", "skillTags", "skillUserProfiles", "trainingGroups", "trainingGroupMembers",
		"contentProfiles", "contentProfileAssignments", "contentProfileAssignmentsHistory", "contentProperties",
		"contentPropertyAssignments", "customContents", "customContentTypes", "customContentTypeFields",
		"digitalSalesRoomTemplates", "digitalSalesRoomTemplateVersions", "externalContentDetails",
		"favoriteStatus", "followStatus", "libraryContents", "libraryContentExpertAssociations", "libraryContentVersions",
		"programs", "programAssociations", "programDates", "programItems", "programRequests", "programRequestDates",
		"programTasks", "programTaskDates", "publishingApprovalWorkflows", "publishingApprovalWorkflowAcknowledgements",
		"publishingApprovalWorkflowStepsHistory", "teamsites", "generatedLivedocs", "generatedLivedocComponents",
		"generatedLivedocFields", "generatedLivedocOutputFormats", "generatedLivedocSlides",
		"livedocCustomContentDataSourceRetrieveDataRequests", "livedocDataSourceInfo",
		"livedocDataSourceRetrieveDataRequests", "livedocGlobalVariableRequests", "digitalSalesRooms",
		"digitalSalesRoomViewingSessions", "livesendComments", "livesendCommentMentions", "livesendCommentReactions",
		"livesendContentViewingSessions", "livesendLinks", "livesendLinkContents", "livesendLinkMeetingContents",
		"livesendLinkMembers", "livesendPageViews", "livesendViewingSessions", "microappScreens", "microappScreenViews",
		"searchClicks", "searchClickMatchDetails", "searchFacets", "searchHistory", "searchWords", "entitlementRoles",
		"externalUsers", "groups", "groupMembers", "indirectGroupManagers", "indirectGroupMembers",
		"users", "userEntitlementRoleAssignments", "userGroupsList", "userProperties", "userPropertyAssignments",
		"workspaceContents", "workspaceContentVersions",
	}

	return components.EndpointRegistryInput{
		providers.ModuleReporting: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
