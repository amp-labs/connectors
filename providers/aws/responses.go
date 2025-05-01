package aws

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsLocation := params.ObjectName

	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath(recordsLocation),
		func(node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("NextToken", "")
		},
		common.GetMarshaledData,
		params.Fields,
	)
}
