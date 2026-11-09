package wordpress

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// nolint:gochecknoglobals
var (
	writeObjects  datautils.Set[string]
	deleteObjects datautils.Set[string]
)

func init() {
	writeOnly := []string{ // supports GET/POST/PATCH
		"sidebars",
	}
	writeAndDelete := []string{ // supports GET/POST/PATCH/DELETE
		"activity-log-events",
		"blocks",
		"comments",
		"feedback",
		"font-families",
		"guidelines",
		"jetpack-forms",
		"jp_pay_order",
		"jp_pay_product",
		"knowledge",
		"media",
		"menu-items",
		"menus",
		"navigation",
		"pages",
		"plugins",
		"posts",
		"tags",
		"template-parts",
		"templates",
		"users",
		"widgets",
		"wp_guideline_type",
		"wp_knowledge_type",
		"wp_pattern_category",
	}

	writeObjects = datautils.NewSetFromList(writeAndDelete)
	writeObjects.Add(writeOnly)

	deleteObjects = datautils.NewSetFromList(writeAndDelete)
}

func (c Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if !writeObjects.Has(params.ObjectName) {
		return nil, fmt.Errorf("%w: object %v", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	var (
		method string
		url    *urlbuilder.URL
		err    error
	)

	if params.IsCreate() {
		method = http.MethodPost
		url, err = c.getWriteDeleteURL(params.ObjectName, "")
	} else {
		method = http.MethodPatch
		url, err = c.getWriteDeleteURL(params.ObjectName, params.RecordId)
	}

	if err != nil {
		return nil, err
	}

	reader, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), reader)
}

func (c Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	identifierField := objectToResponseId.Get(params.ObjectName)

	identifier, err := jsonquery.New(body).TextWithDefault(identifierField, params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success:  true,
			RecordId: identifier,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: identifier,
		Errors:   nil,
		Data:     data,
	}, nil
}
