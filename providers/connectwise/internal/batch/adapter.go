package batch

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

const apiVersion = "v4_6_release/apis/3.0"

// ConnectWise docs state that there must be a limit for the URL length.
// In practice, exceeding the URL size results in "414 Request URI Too Long".
// > We recommend keeping the URL length of each request to a maximum of 2000 characters.
// > This will ensure there are no compatibility issues with various servers and configurations.
const maxURLSize = 1950

// Adapter handles batched operations.
type Adapter struct {
	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	clientID string
}

// NewAdapter creates a new batch Adapter configured to fetch multiple records from ConnectWise API.
func NewAdapter(jsonClient *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, clientID string) *Adapter {
	return &Adapter{
		client:       jsonClient,
		providerInfo: providerInfo,
		clientID:     clientID,
	}
}

func (a *Adapter) clientIdHeader() common.Header {
	return common.Header{
		Key:   "ClientId",
		Value: a.clientID,
	}
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(a.providerInfo.BaseURL, apiVersion, objectPath)
}
