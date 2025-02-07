package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/chilipiper/metadata"
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

		url, err := conn.buildURL(object, metadataPageSize)
		if err != nil {
			metadataResults.Errors[object] = err

			continue
		}

		resp, err := conn.Client.Get(ctx, url)
		if err != nil {
			openAPIFallback(object, &metadataResults)

			continue
		}

		res, err := common.UnmarshalJSON[Response](resp)
		if err != nil {
			openAPIFallback(object, &metadataResults)

			continue
		}

		if len(res.Results) == 0 {
			openAPIFallback(object, &metadataResults)

			continue
		}

		for fld := range res.Results[0] {
			objectMetadata.FieldsMap[fld] = fld
		}

		metadataResults.Result[object] = objectMetadata
	}

	return &metadataResults, nil
}

func metadataFallback(moduleID common.ModuleID, objectName string) (*common.ObjectMetadata, error) {
	metadatResult, err := metadata.Schemas.Select(moduleID, []string{objectName})
	if err != nil {
		return nil, err
	}

	data := metadatResult.Result[objectName]

	return &data, nil
}

func openAPIFallback(obj string, res *common.ListObjectMetadataResult,
) *common.ListObjectMetadataResult { //nolint:unparam
	// Try fallback function
	data, err := metadataFallback(staticschema.RootModuleID, obj)
	if err != nil {
		res.Errors[obj] = err

		return res
	}

	data.DisplayName = obj
	res.Result[obj] = *data

	return res
}
