package dpwrite

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

// ResponseBuilder produces write result via callback. Otherwise, always returns empty result.
type ResponseBuilder struct {
	Build func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error)
}

func (b ResponseBuilder) CreateWriteResult(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
	if b.Build == nil {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	return b.Build(config, body)
}

func (b ResponseBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteResultBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Responder),
	}
}
