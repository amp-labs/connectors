package surveymonkey

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/surveymonkey/metadata"
)

// SurveyMonkey delete API references:
// - DELETE /contacts/{contact_id}
// - DELETE /contact_lists/{contact_list_id}
// - DELETE /surveys/{survey_id}.
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !writeAndDeleteSupportedObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL,
		apiVersion,
		strings.TrimSpace(path),
		params.RecordId,
	)
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

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if httpkit.Status2xx(response.Code) {
		return &common.DeleteResult{Success: true}, nil
	}

	return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
}
