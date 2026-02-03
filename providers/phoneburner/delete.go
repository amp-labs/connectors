package phoneburner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// API reference:
// https://www.phoneburner.com/developer/route_list

func buildDeleteRequest(ctx context.Context, baseURL string, params common.DeleteParams) (*http.Request, error) {
	// All supported deletes are path-ID deletes:
	//   DELETE /rest/1/{object}/{id}
	url, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	_ = ctx
	_ = params
	_ = request

	if err := interpretPhoneBurnerEnvelopeError(response); err != nil {
		return nil, err
	}

	switch response.Code {
	case http.StatusOK, http.StatusAccepted, http.StatusNoContent:
		return &common.DeleteResult{Success: true}, nil
	default:
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}
}

