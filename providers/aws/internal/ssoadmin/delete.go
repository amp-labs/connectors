package ssoadmin

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

func DeleteRequest(
	ctx context.Context, params common.DeleteParams, baseURL, instanceArn string,
) (*http.Request, error) {
	payload := map[string]any{
		"InstanceArn": instanceArn,
	}

	objectProps := Registry[params.ObjectName]
	identifierKey := objectProps.InputRecordID.Delete
	payload[identifierKey] = params.RecordId

	command := core.FormatCommand(ServiceName, objectProps.Commands.Delete)

	return core.BuildRequest(ctx, baseURL, ServiceDomain, ServiceSigningName, command, payload)
}
