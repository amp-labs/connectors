package chorus

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"emails",
		"filters",
		"playlists",
		"scorecards",
		"teams",
		"engagements",
	}

	writeSupport := []string{
		"conversations:validate",
		"conversations:export",
		"join",
		"filters",
		"moments",
		"playlists",
		"smart_playlists",
		"playlists/moments",
		"scorecards:export",
		"video_conferences",
	}

	deleteSupport := []string{
		"conversations",
		"filters",
		"moments",
		"playlists",
		"video_conferences",
	}

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
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
