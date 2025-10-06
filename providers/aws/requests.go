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
	case identitystore.Registry.Has(params.ObjectName):
		return identitystore.ReadRequest(ctx, params, baseURL, c.identityStoreId)
	case ssoadmin.Registry.Has(params.ObjectName):
		return ssoadmin.ReadRequest(ctx, params, baseURL, c.instanceARN)
	default:
		return nil, common.ErrObjectNotSupported
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	baseURL, err := c.getModuleURL()
	if err != nil {
		return nil, err
	}

	switch {
	case identitystore.Registry.Has(params.ObjectName):
		return identitystore.WriteRequest(ctx, params, baseURL, c.identityStoreId)
	case ssoadmin.Registry.Has(params.ObjectName):
		return ssoadmin.WriteRequest(ctx, params, baseURL, c.instanceARN)
	default:
		return nil, common.ErrObjectNotSupported
	}
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	baseURL, err := c.getModuleURL()
	if err != nil {
		return nil, err
	}

	switch {
	case identitystore.Registry.Has(params.ObjectName):
		return identitystore.DeleteRequest(ctx, params, baseURL, c.identityStoreId)
	case ssoadmin.Registry.Has(params.ObjectName):
		return ssoadmin.DeleteRequest(ctx, params, baseURL, c.instanceARN)
	default:
		return nil, common.ErrObjectNotSupported
	}
}
