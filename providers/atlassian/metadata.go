package atlassian

import (
	"context"
	"errors"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"

	"github.com/amp-labs/connectors/common"
)

// ListObjectMetadata lists builtin and custom fields.
// Supports only Issue object. Therefore, objectNames argument is ignored.
// API Reference:
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-fields/#api-rest-api-2-field-get
func (c *Connector) ListObjectMetadata(ctx context.Context, _ []string) (*common.ListObjectMetadataResult, error) {
	url, err := c.getJiraRestApiURL("field")
	if err != nil {
		return nil, err
	}

	rsp, err := c.Clients.JSON.Get(ctx, url.String())
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

	return &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"issue": {
				DisplayName: "Issue",
				FieldsMap:   fields,
			},
		},
		Errors: nil,
	}, nil
}

var (
	ErrParsingMetadata = errors.New("couldn't parse metadata")
	ErrMissingMetadata = errors.New("there is no metadata for object")
)

// Converts API response into the fields' registry.
func (c *Connector) parseFieldsJiraIssue(node *ajson.Node) (map[string]string, error) {
	arr, err := node.GetArray()
	if err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, ErrMissingMetadata
	}

	fieldsMap := make(map[string]string)

	for _, item := range arr {
		name, err := jsonquery.New(item).Str("id", false)
		if err != nil {
			return nil, err
		}

		displayName, err := jsonquery.New(item).Str("name", false)
		if err != nil {
			return nil, err
		}

		fieldsMap[*name] = *displayName
	}

	return fieldsMap, nil
}
