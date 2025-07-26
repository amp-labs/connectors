package atlassian

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
)

// ListObjectMetadata lists builtin and custom fields.
// Supports only Issue object. Therefore, objectNames argument is ignored.
// API Reference:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-fields/#api-rest-api-2-field-get
func (c *Connector) ListObjectMetadata(ctx context.Context, _ []string) (*common.ListObjectMetadataResult, error) {
	url, err := c.getModuleURL("field")
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := rsp.Body()
	if !ok {
		return nil, errors.Join(ErrMissingMetadata, common.ErrEmptyJSONHTTPResponse)
	}

	fields, err := c.parseFieldsJiraIssue(body)
	if err != nil {
		return nil, errors.Join(ErrParsingMetadata, err)
	}

	// Read response is flattened exposing only important fields which happen to not have id.
	// To mitigate this API response the Read method will attach id.
	// Therefore, metadata must include it too.
	fields["id"] = "Id"

	objectMetadata := common.NewObjectMetadata(
		"Issue",
		common.FieldsMetadata{},
	)

	for k, v := range fields {
		objectMetadata.AddField(k, v)
	}

	return &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"issue": *objectMetadata,
		},
		Errors: nil,
	}, nil
}
