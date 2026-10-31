package twenty

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var errDeleteResponse = errors.New("unexpected delete response")

// buildDeleteRequest soft-deletes: DELETE /rest/{namePlural}/{id}?soft_delete=true moves the record to
// Twenty's trash, where it can be restored. soft_delete defaults to false, which destroys the record for good
// (the OpenAPI document the instance serves at /rest/open-api/core: "If false, objects are permanently
// deleted", v2.50.0), so it is always sent. A soft-deleted record leaves the default list; Read returns it
// with ReadParams.Deleted.
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	deleteURL, err := c.recordsURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	deleteURL.WithQueryParam("soft_delete", "true")

	return http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL.String(), nil)
}

// parseDeleteResponse expects data.delete<Object>.id to be the deleted record. Deleting a record that is already
// in the trash, or does not exist, is a 404 ("Record not found"), returned as an error rather than as success:
// Twenty does not tell the two apart. A 2xx without the record's id is an error too.
func (c *Connector) parseDeleteResponse(
	_ context.Context,
	params common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	data, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	keys := data.Keys()
	if len(keys) != 1 {
		return nil, fmt.Errorf("%w: data holds %v", errDeleteResponse, keys)
	}

	id, err := jsonquery.New(data, keys[0]).StringRequired("id")
	if err != nil {
		return nil, err
	}

	if id != params.RecordId {
		return nil, fmt.Errorf("%w: deleted %s, asked for %s", errDeleteResponse, id, params.RecordId)
	}

	return &common.DeleteResult{Success: true}, nil
}
