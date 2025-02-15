package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

// Static file containing openapi spec.
//

//go:embed Users.json
var usersFile []byte

var (
	UsersFileManager = api3.NewOpenapiFileManager(usersFile)
)
