package drift

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	list          = "list"
	object        = "object"
	data          = "data"
	updateAccount = "accounts/update"
)

var recordIdFields = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"accounts":      "accountId",
	"conversations": "id",
	"contacts":      "id",
}, func(k string) string {
	return ""
})

func constructUpdateEndpoint(url *urlbuilder.URL, objectName, recordId string) *urlbuilder.URL {
	switch objectName {
	case accounts:
		url.AddPath("update")
	case users:
		url.AddPath("update")
		url.WithQueryParam("userId", recordId)
	default:
		url.AddPath(recordId)
	}

	return url
}

func constructCreateEndpoint(url *urlbuilder.URL, objectName string) *urlbuilder.URL {
	if objectName == conversations {
		url.AddPath("new")
	}

	if objectName == accounts {
		url.AddPath("create")
	}

	return url
}

func responseSchema(objectName string) (string, string) {
	switch objectName {
	case "users", "conversations", "teams/org", "users/meetings/org":
		return object, data
	case "playbooks", "playbooks/clp":
		return list, ""
	default:
		return object, ""
	}
}

func writeResponseField(objectName string) string {
	switch objectName {
	case "contacts":
		return "data"
	default:
		return ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"users", "conversations", "teams/org", "users/meetings/org",
		"playbooks", "playbooks/clp", "conversations/stats", "scim/Users",
	}

	writeSupport := []string{
		"contacts", "emails/unsubscribe", "contacts/timeline", "conversations", "accounts/create",
		"accounts/update", // updates do not need recordIdPath
		"scim/Users",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			}, {
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
