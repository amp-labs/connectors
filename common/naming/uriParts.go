package naming

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

// APIEndpoints is a set of all endpoints that belong to an API.
// Given the list of endpoints the shortest non-colliding names can be inferred.
type APIEndpoints datautils.StringSet

func NewAPIEndpoints() APIEndpoints {
	return APIEndpoints(datautils.NewStringSet())
}

func (e APIEndpoints) Add(endpoint string) {
	datautils.StringSet(e).AddOne(endpoint)
}

func (e APIEndpoints) ShortestNonCollidingNames() map[string]string {
	registry := make(map[string]*endpointData)

	for endpoint := range e {
		registry[endpoint] = newEndpointData(endpoint)
	}

	hasColliders := true

	for hasColliders {
		hasColliders = false

		groupedColliders := datautils.NamedLists[*endpointData]{}
		for _, endpoint := range registry {
			groupedColliders.Add(endpoint.Name, endpoint)
		}

		for _, colliders := range groupedColliders {
			if len(colliders) > 1 {
				hasColliders = true

				for _, endpoint := range colliders {
					endpoint.ExtendObjectName()
				}
			}
		}
	}

	result := make(map[string]string)

	for endpointURL, data := range registry {
		result[endpointURL] = data.Name
	}

	return result
}

type endpointData struct {
	Full           string
	PrecedingParts []string
	Name           string
}

func newEndpointData(endpoint string) *endpointData {
	parts := strings.Split(endpoint, "/")
	lastIndex := len(parts) - 1

	return &endpointData{
		Full:           endpoint,
		PrecedingParts: parts[:lastIndex],
		Name:           parts[lastIndex],
	}
}

func (d *endpointData) ExtendObjectName() {
	lastIndex := len(d.PrecedingParts)
	d.PrecedingParts = d.PrecedingParts[:lastIndex]
	d.Name = d.PrecedingParts[lastIndex] + "/" + d.Name
}
