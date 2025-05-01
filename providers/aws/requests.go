package aws

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/identitystore"
	"github.com/amp-labs/connectors/providers/aws/internal/ssoadmin"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	baseURL, err := c.getModuleURL()
	if err != nil {
		return nil, err
	}

	switch {
	case identitystore.ReadObjectCommands.Has(params.ObjectName):
		return identitystore.ReadRequest(ctx, params, baseURL, c.identityStoreId)
	case ssoadmin.ReadObjectCommands.Has(params.ObjectName):
		return ssoadmin.ReadRequest(ctx, params, baseURL, c.instanceARN)
	default:
		return nil, common.ErrObjectNotSupported
	}
}
