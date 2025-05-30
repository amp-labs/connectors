package pardot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const MetadataKeyBusinessUnitID = "businessUnitId"

type Adapter struct {
	Client         *common.JSONHTTPClient
	BaseURL        string
	BusinessUnitID string
}

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo, businessUnitID string,
) (*Adapter, error) {
	return &Adapter{
		Client:         client,
		BaseURL:        info.BaseURL,
		BusinessUnitID: businessUnitID,
	}, nil
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, "api/v5/objects", objectName)
}
