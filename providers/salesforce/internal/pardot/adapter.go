package pardot

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	Client         *common.JSONHTTPClient
	BaseURL        string
	BusinessUnitID string
}

var ErrMissingBusinessUnitID = errors.New("missing metadata variable: business unit id")

func NewAdapter(
	client *common.JSONHTTPClient, info *providers.ModuleInfo, metadata map[string]string,
) (*Adapter, error) {

	businessUnitID, ok := metadata["businessUnitId"]
	if !ok || businessUnitID == "" {
		return nil, ErrMissingBusinessUnitID
	}

	return &Adapter{
		Client:         client,
		BaseURL:        info.BaseURL,
		BusinessUnitID: businessUnitID,
	}, nil
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.BaseURL, "api/v5/objects", objectName)
}
