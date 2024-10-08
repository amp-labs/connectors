package deep

import "github.com/amp-labs/connectors/common/urlbuilder"

type URLResolver struct {
	Resolve func(baseURL, objectName string) (*urlbuilder.URL, error)
}
