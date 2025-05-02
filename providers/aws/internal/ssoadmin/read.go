package ssoadmin

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

// nolint:tagliatelle
type readPayload struct {
	*core.ReadPayload
	InstanceArn string `json:"InstanceArn"`
}

func ReadRequest(
	ctx context.Context, params common.ReadParams, baseURL, instanceArn string,
) (*http.Request, error) {
	command := core.FormatCommand(ServiceName, ReadObjectCommands, params.ObjectName)

	return core.BuildRequest(ctx, baseURL, ServiceDomain, ServiceSigningName, command, readPayload{
		ReadPayload: core.NewReadPayload(params),
		InstanceArn: instanceArn,
	})
}
