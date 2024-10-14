package dprequests

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// HeaderSupplements is a list of constant headers that should be attached while performing operation.
// All field includes headers that are shared across all calls.
// This headers will be picked up by deep.Reader, deep.Writer, deep.Remover.
type HeaderSupplements struct {
	All    []common.Header
	Read   []common.Header
	Create []common.Header
	Update []common.Header
	Delete []common.Header
}

func (s HeaderSupplements) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.HeaderSupplements,
		Constructor: handy.PtrReturner(s),
	}
}

func (s HeaderSupplements) ReadHeaders() []common.Header {
	result := make([]common.Header, 0, len(s.All)+len(s.Read))

	result = append(result, s.All...)
	result = append(result, s.Read...)

	return result
}

func (s HeaderSupplements) CreateHeaders() []common.Header {
	result := make([]common.Header, 0, len(s.All)+len(s.Create))

	result = append(result, s.All...)
	result = append(result, s.Create...)

	return result
}

func (s HeaderSupplements) UpdateHeaders() []common.Header {
	result := make([]common.Header, 0, len(s.All)+len(s.Update))

	result = append(result, s.All...)
	result = append(result, s.Update...)

	return result
}

func (s HeaderSupplements) DeleteHeaders() []common.Header {
	result := make([]common.Header, 0, len(s.All)+len(s.Delete))

	result = append(result, s.All...)
	result = append(result, s.Delete...)

	return result
}
