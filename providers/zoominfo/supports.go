package zoominfo

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}

	// Write support is derived from writeObjects rather than restated as a
	// literal, because that map is the real authority: buildWriteRequest cannot
	// build a request without an entry in it. Restating the names here would
	// admit two drift bugs — a registry name with no map entry reaches
	// buildWriteRequest and fails there, and a map entry missing from the
	// registry is silently unreachable. Deriving makes both impossible.
	// Sorted only so the compiled glob is deterministic; match order is
	// irrelevant.
	writeSupport := slices.Sorted(maps.Keys(writeObjects))

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
