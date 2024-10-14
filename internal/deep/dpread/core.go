package dpread

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

var (
	// Implementations.
	_ PaginationStart = FirstPageBuilder{}
	_ PaginationStep  = NextPageBuilder{}
	_ Requester       = RequestGet{}
)

// PaginationStart is connector component, which alters URL with query parameters to prepare for the first page call.
type PaginationStart interface {
	requirements.ConnectorComponent

	FirstPage(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
}

// PaginationStep is connector component, which creates next page token based on current URL and response.
type PaginationStep interface {
	requirements.ConnectorComponent

	NextPage(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error)
}

// Requester is connector component, which selects appropriate HTTP method for an object.
// In most cases GET operation is sufficient. However, the implementation could return POST to perform a search.
type Requester interface {
	requirements.ConnectorComponent

	MakeReadRequest(objectName string, clients dprequests.Clients) (common.ReadMethod, []common.Header)
}

// Responder is connector component, which produces response parser which will extract desired Objects.
type Responder interface {
	requirements.ConnectorComponent

	GetRecordsFunc(config common.ReadParams) (common.RecordsFunc, error)
}
