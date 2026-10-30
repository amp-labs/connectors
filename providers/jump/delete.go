package jump

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// buildDeleteRequest calls the object's delete mutation with the record ID.
// https://my.jumpapp.com/enterprise/documentation/reference?tab=mutations
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	mutations, ok := writeMutations[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return c.buildMutationRequest(ctx, mutations.delete, map[string]any{
		"id": params.RecordId,
		// Account-scoped API keys must name the user the mutation acts for.
		"userId": c.userID,
	})
}

// parseDeleteResponse checks the "success" flag of the delete mutation.
//
//	{
//	  "data": {
//	    "deleteContact": {
//	      "success": true
//	    }
//	  }
//	}
func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	mutation := writeMutations[params.ObjectName].delete

	if err := interpretGraphQLError(resp, mutation); err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	success, err := jsonquery.New(body, "data", mutation).BoolRequired("success")
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, fmt.Errorf("%w: %s returned success false", common.ErrRequestFailed, mutation)
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
