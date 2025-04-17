package ashby

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/ashby/metadata"
)

var (
	// nolint:gochecknoglobals,lll
	supportPagination = datautils.NewSet("application.list", "applicationFeedback.list", "candidate.list",
		"candidateTag.list", "customField.list", "feedbackFormDefinition.list", "interview.list", "interviewSchedule.list",
		"feedbackFormDefinition.list", "interviewerPool.list", "job.list", "jobTemplate.list", "offer.list", "opening.list",
		"project.list", "surveyFormDefinition.list", "user.list")

	//nolint:gochecknoglobals
	supportSince = datautils.NewSet("application.list", "applicationFeedback.list", "interviewSchedule.list")
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	//nolint:lll
	writeSupport := []string{
		"application.create", "candidate.create", "candidate.createNote", "candidate.addTag",
		"candidate.addProject", "candidateTag.create", "customField.create", "department.create",
		"hiringTeam.addMember", "interviewSchedule.create", "interviewerPool.create", "interviewerPool.addUser",
		"job.create", "location.create", "offer.create", "offerProcess.start", "opening.create", "referral.create",
		"surveyRequest.create", "surveySubmission.create", "webhook.create",

		"application.change_source", "application.change_stage", "application.transfer", "application.update",
		"application.updateHistory", "application.addHiringTeamMember", "application.removeHiringTeamMember",
		"applicationFeedback.submit", "applicationForm.submit", "approvalDefinition.update", "assessment.addCompletedToCandidate",
		"assessment.start", "assessment.update", "assessment.cancel", "candidate.anonymize", "candidate.update", "customField.setValue",
		"hiringTeam.addMember", "interviewSchedule.cancel", "interviewSchedule.update", "interviewerPool.update", "interviewerPool.archive",
		"interviewerPool.restore", "interviewerPool.addUser", "interviewerPool.removeUser", "job.setStatus", "job.update", "jobPosting.update",
		"offer.start", "opening.addJob", "opening.removeJob", "opening.addLocation", "opening.removeLocation", "opening.setOpeningState", "opening.setArchived",
		"opening.update", "webhook.update",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
