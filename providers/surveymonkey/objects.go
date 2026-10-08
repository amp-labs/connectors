package surveymonkey

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing (match schemas.json keys).
const (
	objectBenchmarkBundles      = "benchmark_bundles"
	objectContactFields         = "contact_fields"
	objectContactLists          = "contact_lists"
	objectContacts              = "contacts"
	objectGroups                = "groups"
	objectOrganizations         = "organizations"
	objectQuestionBankQuestions = "question_bank_questions"
	objectRoles                 = "roles"
	objectSurveyCategories      = "survey_categories"
	objectSurveyFolders         = "survey_folders"
	objectSurveyLanguages       = "survey_languages"
	objectSurveys               = "surveys"
	objectSurveyTemplates       = "survey_templates"
	objectTeamSurveyTemplates   = "team_survey_templates"
	objectWorkgroups            = "workgroups"
)

//nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectBenchmarkBundles,
	objectContactFields,
	objectContactLists,
	objectContacts,
	objectGroups,
	objectOrganizations,
	objectQuestionBankQuestions,
	objectRoles,
	objectSurveyCategories,
	objectSurveyFolders,
	objectSurveyLanguages,
	objectSurveys,
	objectSurveyTemplates,
	objectTeamSurveyTemplates,
	objectWorkgroups,
)
