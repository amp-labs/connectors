package batch

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1.0"

type Strategy struct {
	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
}

func NewStrategy(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo) *Strategy {
	return &Strategy{
		client:       client,
		providerInfo: providerInfo,
	}
}

// https://learn.microsoft.com/en-us/graph/json-batching?tabs=http#creating-a-batch-request
func (s Strategy) getBatchURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, apiVersion, "$batch")
}

func (s Strategy) getVersionedRootURL() string {
	return s.providerInfo.BaseURL + "/" + apiVersion
}
