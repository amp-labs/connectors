package paramsbuilder

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
)

type ModuleHolder interface {
	GiveModule() *Module
}

// Module represents a sub-product of a provider.
// This is relevant where there are several APIs for different sub-products, and the APIs
// are versioned differently or have different ways of constructing URLs and requests for reading/writing.
type Module struct {
	Selection common.Module
	// Connector implementation must provide list of modules that are currently supported.
	// This defines a list of modules a user could switch to.
	// Validation will check module identifiers are consistent.
	supported common.Modules
}

func (p *Module) GiveModule() *Module {
	return p
}

func (p *Module) ValidateParams() error {
	// Module ID must match the ID which is used to index modules.
	for id, module := range p.supported {
		if module.ID != id {
			return common.ErrInvalidModuleDeclaration
		}
	}

	return nil
}

// WithModule allows API module selection.
// Connector implementation must provide list of modules that are currently supported.
// This defines a list of module a user could switch to. Any out of scope module will be ignored.
// If user supplied unsupported module, then it will use this fallback.
func (p *Module) WithModule(moduleID common.ModuleID, supported common.Modules, fallbackID common.ModuleID) {
	modules := handy.NewDefaultMap(supported,
		func(unknownID common.ModuleID) common.Module {
			// Unknown module ids will receive a fallback module.
			return supported[fallbackID]
		},
	)

	p.Selection = modules.Get(moduleID)
	p.supported = supported
}
