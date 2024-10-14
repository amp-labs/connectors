package deep

import (
	"github.com/amp-labs/connectors/internal/deep/dprequests"
)

// Clients is a major connector component which provides HTTP client functionality.
// Embed this into connector struct.
type Clients = dprequests.Clients
