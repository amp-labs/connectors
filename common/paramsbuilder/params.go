// Package paramsbuilder provides common parameters used to initialize connectors.
// Implementor would pick every relevant parameter and use them to compose one unified parameters list.
// Then methods of format "With<some-property-name>" should be configured to connector needs
// and exposed to end user via delegation. Most would do delegation only.
package paramsbuilder

import "github.com/amp-labs/connectors/common"

// ParamAssurance checks that param data is valid
// Every param instance must implement it.
type ParamAssurance interface {
	ValidateParams() error
}

// Apply will apply options to construct a ready to go set of parameters.
// This is a generalized constructor of parameters used to initialize any connector.
// To qualify as a parameter one must have data validation laid out.
func Apply[P ParamAssurance](params P, opts []func(params *P)) (paramsOut *P, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		// Options may have calls to panic().
		// We allow this and recover by converting it to normal error.
		outErr = cause
		paramsOut = nil
	})

	for _, opt := range opts {
		opt(&params)
	}

	return &params, params.ValidateParams()
}
