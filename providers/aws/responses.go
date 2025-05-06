package aws

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
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

func (c *Connector) parseWriteResponse(
	ctx context.Context, params common.WriteParams, request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	var outputRecordID core.OutputRecordID

	switch {
	case identitystore.Registry.Has(params.ObjectName):
		outputRecordID = identitystore.Registry[params.ObjectName].OutputRecordID
	case ssoadmin.Registry.Has(params.ObjectName):
		outputRecordID = ssoadmin.Registry[params.ObjectName].OutputRecordID
	}

	var recordID string
	if len(params.RecordId) == 0 {
		recordID = outputRecordID.Create.Extract(body, params.RecordId)
	} else {
		recordID = outputRecordID.Update.Extract(body, params.RecordId)
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context, params common.DeleteParams, request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
