package deep

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type HeaderSupplements struct {
	All    []common.Header
	Read   []common.Header
	Create []common.Header
	Update []common.Header
	Delete []common.Header
}

func (s HeaderSupplements) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "headerSupplements",
		Constructor: handy.Returner(s),
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
