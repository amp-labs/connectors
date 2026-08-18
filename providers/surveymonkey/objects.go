package surveymonkey

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing (match schemas.json keys).
const (
	objectBenchmarkBundles      = "benchmark_bundles"
	objectContactLists          = "contact_lists"
	objectContacts              = "contacts"
	objectGroups                = "groups"
	objectOrganizations         = "organizations"
	objectQuestionBankQuestions = "question_bank_questions"
	objectRoles                 = "roles"
	objectSurveyCategories      = "survey_categories"
	objectSurveyFolders         = "survey_folders"
	objectSurveyLanguages       = "survey_languages"
	objectSurveyTemplates       = "survey_templates"
	objectTeamSurveyTemplates   = "team_survey_templates"
)

//nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectBenchmarkBundles,
	objectContactLists,
	objectContacts,
	objectGroups,
	objectOrganizations,
	objectQuestionBankQuestions,
	objectRoles,
	objectSurveyCategories,
	objectSurveyFolders,
	objectSurveyLanguages,
	objectSurveyTemplates,
	objectTeamSurveyTemplates,
)
