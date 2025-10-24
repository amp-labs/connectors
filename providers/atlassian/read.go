package atlassian

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	// issues API support upto 500 issues per API call.
	pageSize = 200
	issues   = "issues"
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
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL()
	if err != nil {
		return nil, err
	}

	if config.ObjectName == issues {
		var minutes int64

		write := c.Client.Post

		timeDuration := time.Since(time.Unix(0, 0).UTC())
		minutes = int64(timeDuration.Minutes())

		if !config.Since.IsZero() {
			diff := time.Since(config.Since)
			minutes = int64(diff.Minutes())
		}

		reqBody := issueRequest{
			Fields:     config.Fields.List(),
			JQL:        fmt.Sprintf(`updated > "-%vm"`, minutes),
			MaxResults: pageSize,
		}

		if len(config.NextPage) > 0 {
			reqBody.NextPageToken = config.NextPage.String()
		}

		resp, err := write(ctx, url.String(), reqBody)
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

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecords,
		common.MakeMarshaledDataFunc(flattenRecord),
		config.Fields,
	)
}

func (c *Connector) buildReadURL() (*urlbuilder.URL, error) {
	url, err := c.getModuleURL("search/jql")
	if err != nil {
		return nil, err
	}

	return url, nil
}
