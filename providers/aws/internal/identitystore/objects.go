// nolint:gochecknoglobals
package identitystore

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	ServiceName        = "AWSIdentityStore"
	ServiceDomain      = "identitystore"
	ServiceSigningName = "identitystore"
)

var ReadObjectCommands = datautils.Map[string, string]{
	"Groups": "ListGroups",
	"Users":  "ListUsers",
}
