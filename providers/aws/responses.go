package aws

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws/internal/identitystore"
	"github.com/amp-labs/connectors/providers/aws/internal/ssoadmin"
	"github.com/spyzhov/ajson"
)

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsLocation := getReadRecordsLocation(params)

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

func getReadRecordsLocation(params common.ReadParams) string {
	recordsLocation, ok := identitystore.Schemas.FindArrayFieldName(providers.ModuleAWSIdentityCenter, params.ObjectName)
	if ok {
		return recordsLocation
	}

	// Object must be coming from this service.
	recordsLocation, _ = ssoadmin.Schemas.FindArrayFieldName(providers.ModuleAWSIdentityCenter, params.ObjectName)

	return recordsLocation
}
