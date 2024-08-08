package paramsbuilder

import (
	"errors"
	"fmt"
)

var ErrNoSupportedModule = errors.New("no supported module was chosen")

// Module represents a sub-product of a provider.
// This is relevant where there are several APIs for different sub-products, and the APIs
// are versioned differently or have different ways of constructing URLs and requests for reading/writing.
type Module struct {
	// Name represents the name of the sub-product. e.g. crm, marketing, etc.
	Name string
	// Connector implementation must provide list of modules that are currently supported.
	// This defines a list of module a user could switch to. Any out of scope module will be ignored.
	supported []APIModule
	// If user supplied unsupported module it will use this fallback.
	fallback *APIModule
}

func (p *Module) ValidateParams() error {
	// making sure the provided module is supported.
	// If the provided module is not supported, use fallback.
	if !p.isSupported() {
		if p.fallback == nil {
			// not supported and user didn't provide a fallback
			return ErrNoSupportedModule
		}

		// replace with fallback module
		p.Name = p.fallback.String()

		// even fallback is not supported
		if !p.isSupported() {
			return ErrNoSupportedModule
		}
	}

	return nil
}

func (p *Module) WithModule(module APIModule, supported []APIModule, defaultModule *APIModule) {
	p.Name = module.String()
	p.supported = supported
	p.fallback = defaultModule
}

func (p *Module) isSupported() bool {
	for _, mod := range p.supported {
		if p.Name == mod.String() {
			return true
		}
	}

	return false
}

type APIModule struct {
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

func (a APIModule) String() string {
	if len(a.Label) == 0 {
		return a.Version
	}

	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}
