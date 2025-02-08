package googledicsovery

import "github.com/amp-labs/connectors/tools/fileconv/api3"

// WithDisplayNamePostProcessors will apply processors in the given order.
func WithDisplayNamePostProcessors(processors ...api3.DisplayNameProcessor) Option {
	return func(params *parameters) {
		params.displayPostProcessing = func(displayName string) string {
			for _, processor := range processors {
				displayName = processor(displayName)
			}

			return displayName
		}
	}
}

type parameters struct {
	displayPostProcessing api3.DisplayNameProcessor
}

type Option = func(params *parameters)

func createParams(opts []Option) *parameters {
	params := parameters{
		// Default values are setup here.
		displayPostProcessing: func(displayName string) string {
			return displayName
		},
	}
	for _, opt := range opts {
		opt(&params)
	}

	return &params
}
