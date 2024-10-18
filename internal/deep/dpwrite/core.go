package dpwrite

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

var (
	// Implementations.
	_ Requester = RequestPostPut{}
	_ Requester = RequestPostNoop{}
	_ Requester = RequestPostPatch{}
	_ Requester = RequestPostPost{}
	_ Responder = ResponseBuilder{}
)

type Requester interface {
	requirements.ConnectorComponent

	MakeCreateRequest(
		objectName string, url *urlbuilder.URL, clients dprequests.Clients,
	) (common.WriteMethod, []common.Header)
	MakeUpdateRequest(
		objectName, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
	) (common.WriteMethod, []common.Header)
}

// Responder is connector component, which parses and produces write result.
type Responder interface {
	requirements.ConnectorComponent

	CreateWriteResult(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error)
}
