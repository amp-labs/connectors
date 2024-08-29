package apollo

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// ListObjectMetadata creates metadata of object via reading objects using Apollo API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		url, err := c.getAPIURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		// Check nil response body, to avoid panic.
		if resp == nil || resp.Body == nil {
			metadataResult.Errors[obj] = common.ErrEmptyResponse

			continue
		}

		metadata, err := parseMetadataFromResponse(resp, obj)
		if err != nil {
			if errors.Is(err, common.ErrMetadataLoadFailure) {
				metadataResult.Errors[obj] = common.ErrEmptyResponse
			} else {
				return nil, err
			}
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = metadata
	}

	return &metadataResult, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse, obj string) (common.ObjectMetadata, error) {
	bb := resp.Body.Source()
	if bb == nil {
		return common.ObjectMetadata{}, common.ErrEmptyResponse
	}

	metadata := common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	root, err := ajson.Unmarshal(bb)
	if err != nil {
		return common.ObjectMetadata{}, err
	}

	resultArr, err := jsonquery.New(root).Array(obj, true)
	if err != nil {
		return common.ObjectMetadata{}, err
	}

	objectResponnse := resultArr[0].MustObject()

	// Using the first result data to generate the metadata.
	for k := range objectResponnse {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}
