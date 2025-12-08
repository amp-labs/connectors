package pipedrive

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/internal/crm"
	"github.com/amp-labs/connectors/providers/pipedrive/internal/legacy"
)

// Connector represents the Pipedrive Connector.
type Connector struct {
	Client *common.JSONHTTPClient

	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID

	crmAdapter    *crm.Adapter
	legacyAdapter *legacy.Adapter
}

// NewConnector constructs the Pipedrive Connector and returns it, Fails
// if any of the required fields are not instantiated.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ModulePipedriveLegacy),
	)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	conn.providerInfo, err = providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	conn.SetBaseURL(conn.moduleInfo.BaseURL)

	conn.moduleID = params.Selection.ID

	if conn.moduleID == providers.ModulePipedriveCRM {
		conn.crmAdapter = crm.NewAdapter(conn.Client, conn.moduleInfo)
	}

	if conn.moduleID == providers.ModulePipedriveLegacy {
		conn.legacyAdapter = legacy.NewAdapter(conn.Client, conn.moduleInfo)
	}

	return conn, nil
}

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	return c.legacyAdapter.Read(ctx, config)
}

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	return c.legacyAdapter.Write(ctx, config)
}

// Provider returns the pipedrive provider instance.
func (c *Connector) Provider() providers.Provider {
	return providers.Pipedrive
}

// String implements the fmt.Stringer interface.
func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) SetBaseURL(newURL string) {
	c.providerInfo.BaseURL = newURL
	c.moduleInfo.BaseURL = newURL
	c.HTTPClient().Base = newURL
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.ListObjectMetadata(ctx, objectNames)
	}

	return c.legacyAdapter.ListObjectMetadata(ctx, objectNames)
}
