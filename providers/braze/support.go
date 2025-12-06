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

type pathConfig struct {
	PrimaryKey string
	CreatePath string
	UpdatePath string

	// GenericPath is only filled, incase it's the only path for such object,
	// it's has either create, update or delete only, not both.
	GenericPath string
}

var writePaths = map[string]pathConfig{ //nolint: gochecknoglobals
	"templates/email": {
		PrimaryKey: "email_template_id",
		UpdatePath: "templates/email/update",
		CreatePath: "templates/email/create",
	},
	"messages/schedules": {
		PrimaryKey: "schedule_id",
		UpdatePath: "messages/schedule/update",
		CreatePath: "messages/schedule/update/create",
	},
	"content_blocks": {
		PrimaryKey: "content_block_id",
		UpdatePath: "content_blocks/update",
		CreatePath: "content_blocks/create",
	},
	"campaigns/triggers/schedules": {
		PrimaryKey: "schedule_id",
		UpdatePath: "campaigns/trigger/schedule/update",
		CreatePath: "campaigns/trigger/schedule/update",
	},
	"live_activities": {
		PrimaryKey: "activity_id",
		UpdatePath: "messages/live_activity/update",
		CreatePath: "messages/live_activity/start",
	},
	"user/alias": {
		PrimaryKey: "alias_updates",
		CreatePath: "users/alias/update",
		UpdatePath: "users/alias/new",
	},
	"canvas/triggers/schedules": {
		PrimaryKey: "schedule_id",
		UpdatePath: "canvas/trigger/schedule/update",
		CreatePath: "canvas/trigger/schedule/create",
	},
	"email/bounce": {
		GenericPath: "email/bounce/remove",
	},
	"subscription/status": {
		GenericPath: "subscription/status/set",
	},
	"preference center": {
		GenericPath: "preference_center/v1",
	},
	"send/id": {
		GenericPath: "sends/id/create",
	},
	"messages": {
		GenericPath: "messages/send",
	},
}

func getPath(objectName string, recordData any) (string, error) {
	paths, exists := writePaths[objectName]
	if !exists {
		return objectName, nil
	}

	if paths.GenericPath != "" {
		return paths.GenericPath, nil
	}

	recordDataMap, ok := recordData.(map[string]any)
	if !ok {
		return "", ErrInvalidData
	}

	if _, exists := recordDataMap[paths.PrimaryKey]; exists {
		return paths.UpdatePath, nil
	}

	return paths.CreatePath, nil
}
