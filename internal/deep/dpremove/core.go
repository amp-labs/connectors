package dpremove

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var (
	// Implementations.
	_ Requester = RequestDelete{}
)

// Requester is connector component, which selects appropriate HTTP method for object removal.
type Requester interface {
	requirements.ConnectorComponent

	MakeDeleteRequest(objectName, recordID string, clients dprequests.Clients) (common.DeleteMethod, []common.Header)
}

