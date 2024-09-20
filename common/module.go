package common

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/handy"
)

// ErrInvalidModuleDeclaration occurs when the identifier used for map indexing in Modules
// does not match the module's ID.
var ErrInvalidModuleDeclaration = errors.New("supported modules are not correctly defined")

type ModuleID string

type Modules = handy.Map[ModuleID, Module]

// Module represents a set of endpoints and functionality available by provider.
// Single provider may support multiple modules, requiring the user to choose a module before making requests.
// Modules may differ by version or by theme, covering different systems or functionalities.
type Module struct {
	ID      ModuleID
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

func (a Module) Path() string {
	if len(a.Label) == 0 {
		return a.Version
	}

	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}
