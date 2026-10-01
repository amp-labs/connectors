package jira

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/atlassian/internal/jql"
)

const (
	// issues API support upto 500 issues per API call.
	pageSize = 200
)

type issueRequest struct {
	Fields        []string `json:"fields"`
	FieldsByKeys  bool     `json:"fieldsByKeys,omitempty"`
	JQL           string   `json:"jql"`
	MaxResults    int      `json:"maxResults"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

// Read only returns a list of Jira Issues.
// You can provide the following values:
// * ObjectName - ignored.
// * NextPage - to get next page which may have no elements left.
// * Since - to scope the time frame, precision is in minutes.
func (a *Adapter) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	switch config.ObjectName {
	case "issue":
	case "issues":
	default:
		logging.Logger(ctx).Warn(
			"using Atlassian connector with unknown object", "objectName", config.ObjectName)
	}

	url, err := a.getSearchIssuesURL()
	if err != nil {
		return nil, err
	}

	jqlQuery := jql.New().
		SinceMinutes(config.Since).
		UntilMinutes(config.Until).
		String()

	reqBody := issueRequest{
		Fields:     config.Fields.List(),
		JQL:        jqlQuery,
		MaxResults: pageSize,
	}

	if len(config.NextPage) > 0 {
		reqBody.NextPageToken = config.NextPage.String()
	}

	resp, err := a.JSONHTTPClient().Post(ctx, url.String(), reqBody)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		getRecords,
		getNextRecordIssues,
		common.MakeMarshaledDataFunc(flattenRecord),
		config.Fields,
	)
}

func (a *Adapter) getSearchIssuesURL() (*urlbuilder.URL, error) {
	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-search-jql-post
	return a.getModuleURL("search/jql")
}
