package identitystore

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

func WriteRequest(
	ctx context.Context, params common.WriteParams, baseURL, identityStoreID string,
) (*http.Request, error) {
	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	recordData["IdentityStoreId"] = identityStoreID

	var command core.Command

	if len(params.RecordId) == 0 {
		objectProps := Registry[params.ObjectName]
		command = core.FormatCommand(ServiceName, objectProps.Commands.Create)
	} else {
		objectProps := Registry[params.ObjectName]
		command = core.FormatCommand(ServiceName, objectProps.Commands.Update)

		identifierKey := objectProps.InputRecordID.Update
		recordData[identifierKey] = params.RecordId
	}

	return core.BuildRequest(ctx, baseURL, ServiceDomain, ServiceSigningName, command, recordData)
}
