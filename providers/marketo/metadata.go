package marketo

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/marketo/metadata"
)

type responseObject struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
	// Other fields
}

// ListObjectMetadata creates metadata of object via reading objects using Marketo API.
// If it fails to fretrieve the metadata, It tried using static schema file in metadata dir.
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
		// Constructing the request url.
		url, err := c.getAPIURL(obj)
		if err != nil {
			return nil, err
		}

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			runFallback(obj, &metadataResult)

			continue
		}

		if _, ok := resp.Body(); !ok {
			runFallback(obj, &metadataResult)

			continue
		}

		metadata, err := parseMetadataFromResponse(resp)
		if err != nil {
			if errors.Is(err, common.ErrMetadataLoadFailure) {
				runFallback(obj, &metadataResult)

				continue
			} else {
				return nil, err
			}
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = *metadata
	}

	return &metadataResult, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMetadataLoadFailure
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}

func metadataFallback(objectName string) (*common.ObjectMetadata, error) {
	schemas, err := metadata.FileManager.LoadSchemas()
	if err != nil {
		return nil, common.ErrMetadataLoadFailure
	}

	metadatResult, err := schemas.Select([]string{objectName})
	if err != nil {
		return nil, err
	}

	metadata := metadatResult.Result[objectName]

	return &metadata, nil
}

func runFallback(obj string, res *common.ListObjectMetadataResult) *common.ListObjectMetadataResult { //nolint:unparam
	// Try fallback function
	metadata, err := metadataFallback(obj)
	if err != nil {
		res.Errors[obj] = err

		return res
	}

	metadata.DisplayName = obj
	res.Result[obj] = *metadata

	return res
}
