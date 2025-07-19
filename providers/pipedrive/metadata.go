package pipedrive

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/pipedrive/metadata"
)

type responseData struct {
	Data    []map[string]any `json:"data"`
	Success bool             `json:"success"`
	// Other fields unrelated to metadata generation.
}

// ListObjectMetadata returns metadata for an object by sampling an object from Pipedrive's API.
// If that fails, it generates object metadata by parsing Pipedrive's OpenAPI files.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		// Constructing the  request url.
		url, err := c.getAPIURL(obj)
		if err != nil {
			return nil, err
		}

		// Limiting the response data to 1 record.
		// we only use 1 record for the metadata generation.
		// no need to query several records.
		url.WithQueryParam(limitQuery, "1")

		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		data, err := metadataMapper(res, c.Module.ID, obj)
		if err != nil {
			return nil, err
		}

		objMetadata.Result[obj] = *data
	}

	return &objMetadata, nil
}

// metadataMapper constructs the metadata fields to a new map and returns it.
// Returns an error if it faces any in unmarshalling the response.
func metadataMapper(
	resp *common.JSONHTTPResponse, moduleID common.ModuleID, obj string,
) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseData](resp)
	if err != nil {
		return nil, err
	}

	mdt, err := metadata.Schemas.SelectOne(moduleID, obj)
	if err != nil {
		return nil, err
	}

	// Ensure the response data array, has at least 1 record.
	// If there is no data, we use only the static schema file.
	if len(response.Data) == 0 {
		return mdt, nil
	}

	// Looping on the first index of the response data.
	fields := response.Data[0]
	for fld := range fields {
		mdt.AddField(fld, fld)
	}

	return mdt, nil
}
