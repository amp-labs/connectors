package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
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

		if !fetchDataFields(ctx, conn, object, &objectMetadata, &metadataResults) {
			openAPIFallback(object, &metadataResults)

			continue
		}
	}

	return &metadataResults, nil
}

func fetchDataFields(
	ctx context.Context,
	conn *Connector, obj string,
	mtd *common.ObjectMetadata,
	res *common.ListObjectMetadataResult,
) bool {
	url, err := conn.buildURL(obj, metadataPageSize)
	if err != nil {
		return false
	}

	jsonResp, err := conn.Client.Get(ctx, url)
	if err != nil {
		return false
	}

	resp, err := common.UnmarshalJSON[Response](jsonResp)
	if err != nil {
		return false
	}

	if len(resp.Results) == 0 {
		return false
	}

	for fld := range resp.Results[0] {
		// TODO fix deprecated
		mtd.FieldsMap[fld] = fld // nolint:staticcheck
	}

	res.Result[obj] = *mtd

	return true
}

func metadataFallback(moduleID common.ModuleID, objectName string) (*common.ObjectMetadata, error) {
	metadataResult, err := metadata.Schemas.Select(moduleID, []string{objectName})
	if err != nil {
		return nil, err
	}

	data, exists := metadataResult.Result[objectName]
	if !exists {
		return nil, metadataResult.Errors[objectName]
	}

	return &data, nil
}

func openAPIFallback(obj string, res *common.ListObjectMetadataResult,
) *common.ListObjectMetadataResult { //nolint:unparam
	// Try fallback function
	data, err := metadataFallback(common.ModuleRoot, obj)
	if err != nil {
		res.Errors[obj] = err

		return res
	}

	data.DisplayName = obj
	res.Result[obj] = *data

	return res
}
