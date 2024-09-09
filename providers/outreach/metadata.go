package outreach

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

type Data struct {
	Data []DataItem `json:"data"`
}

type DataItem struct {
	Type          string         `json:"type"`
	ID            int            `json:"id"`
	Relationships map[string]any `json:"relationships"`
	Attributes    map[string]any `json:"attributes"`
}

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
		url, err := c.getApiURL(obj)
		if err != nil {
			return nil, err
		}

		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			objMetadata.Errors[obj] = err

			continue
		}

		metadata, err := metadataMapper(res)
		if err != nil {
			return nil, err
		}

		metadata.DisplayName = obj
		objMetadata.Result[obj] = *metadata
	}

	return &objMetadata, nil
}

func metadataMapper(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[Data](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	attributes := response.Data[0].Attributes
	for k := range attributes {
		metadata.FieldsMap[k] = k
	}

	// Append id in the metadata response. Only adds it, if available.
	// 0 is not a valid id in outreach types. Id are read-only and starts at 1.
	if response.Data[0].ID != 0 {
		metadata.FieldsMap[idKey] = idKey
	}

	return metadata, nil
}
