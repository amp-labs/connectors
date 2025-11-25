package braze

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}
	writeSupport := []string{"*"}

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

var readEndpointsByObject = map[string]string{ //nolint: gochecknoglobals
	"campaigns":           "campaigns/list",
	"canvas":              "canvas/list",
	"segments":            "segments/list",
	"preference_center":   "preference_center/v1/list",
	"subscription/status": "subscription/status/get",
	"content_blocks":      "content_blocks/list",
	"templates/email":     "templates/email/list",
	"purchases":           "purchases/product_list ",
}

func constructWritePath(params common.WriteParams) (string, error) { //nolint: cyclop,funlen
	recordData, ok := params.RecordData.(map[string]any)
	if !ok {
		return "", ErrInvalidData
	}

	switch params.ObjectName {
	case "content_blocks":
		if _, exists := recordData["content_block_id"]; exists {
			// this means we're updating so we append /update
			return "content_blocks/update", nil
		}

		return "content_blocks/create", nil

	case "templates/email":
		if _, exists := recordData["email_template_id"]; exists {
			// this means we're updating so we append /update
			return "templates/email/update", nil
		}

		return "templates/email/create", nil

	case "messages/schedules":
		if _, exists := recordData["schedule_id"]; exists {
			// this means we're updating so we append /update
			return "messages/schedule/update", nil
		}

		return "messages/schedule/update/create", nil

	case "campaigns/triggers/schedules":
		if _, exists := recordData["schedule_id"]; exists {
			// this means we're updating so we append /update
			return "campaigns/trigger/schedule/update", nil
		}

		return "campaigns/trigger/schedule/create", nil
	case "live activities":
		if _, exists := recordData["activity_id"]; exists {
			// this means we're updating so we append /update
			return "messages/live_activity/update", nil
		}

		return "messages/live_activity/start", nil

	case "user/alias":
		if _, exists := recordData["alias_updates"]; exists {
			// this means we're updating so we append /update
			return "users/alias/update", nil
		}

		return "users/alias/new", nil

	case "canvas/triggers/schedules":
		if _, exists := recordData["schedule_id"]; exists {
			return "canvas/trigger/schedule/update", nil
		}

		return "canvas/trigger/schedule/create", nil

	case "bounced emails":
		return "email/bounce/remove", nil

	case "subscription/status":
		return "subscription/status/set", nil

	case "preference center":
		return "preference_center/v1", nil

	case "send ids":
		return "sends/id/create", nil

	case "messages":
		return "sends/id/create", nil

	default:
		return params.ObjectName, nil
	}
}
