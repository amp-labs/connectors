package paramsbuilder

import (
	"errors"
	"fmt"
)

var ErrNoSupportedModule = errors.New("no supported module was chosen")

// Module params adds suffix to URL controlling API versions.
// This is relevant where there are several APIs for different product areas or sub-products, and the APIs
// are versioned differently or have different ways of constructing URLs from object names.
type Module struct {
	// Suffix represents part of URL string capturing the concept of module.
	Suffix string
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
		p.Suffix = p.fallback.String()

		// even fallback is not supported
		if !p.isSupported() {
			return ErrNoSupportedModule
		}
	}

	return nil
}

func (p *Module) WithModule(module APIModule, supported []APIModule, defaultModule *APIModule) {
	p.Suffix = module.String()
	p.supported = supported
	p.fallback = defaultModule
}

func (p *Module) isSupported() bool {
	for _, mod := range p.supported {
		if p.Suffix == mod.String() {
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
