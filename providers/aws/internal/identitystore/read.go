package identitystore

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

// nolint:tagliatelle
type readPayload struct {
	*core.ReadPayload

	IdentityStoreID string `json:"IdentityStoreId"`
}

func ReadRequest(
	ctx context.Context, params common.ReadParams, baseURL, identityStoreID string,
) (*http.Request, error) {
	objectProps := Registry[params.ObjectName]

	command := core.FormatCommand(ServiceName, objectProps.Commands.Read)

	return core.BuildRequest(ctx, baseURL, ServiceDomain, ServiceSigningName, command, readPayload{
		ReadPayload:     core.NewReadPayload(params),
		IdentityStoreID: identityStoreID,
	})
}
