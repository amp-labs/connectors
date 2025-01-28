package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// Sample OK Response
// {
// 	"results": [
// 		...
// 	],
// 	"total": 0,
// 	"page": 0,
// 	"pageSize": 0
// }

type Response struct {
	Results []map[string]any `json:"results"`
	// The rest of the fields
}

// ListObjectMetadata creates metadata of objects via reading objects using ChiliPiper API.
// If fails uses the OpenAPI specification files.
func (conn *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResults := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, object := range objectNames {
		objectMetadata := common.ObjectMetadata{
			FieldsMap:   make(map[string]string),
			DisplayName: naming.CapitalizeFirstLetterEveryWord(object),
		}

		url, err := conn.buildReadURL(object)
		if err != nil {
			return nil, err
		}

		resp, err := conn.Client.Get(ctx, url.String())
		if err != nil {
			metadataResults.Errors[object] = err

			continue
		}

		res, err := common.UnmarshalJSON[Response](resp)
		if err != nil {
			metadataResults.Errors[object] = err

			continue
		}

		for fld := range res.Results[0] {
			objectMetadata.FieldsMap[fld] = fld
		}

		metadataResults.Result[object] = objectMetadata
	}

	return &metadataResults, nil
}
