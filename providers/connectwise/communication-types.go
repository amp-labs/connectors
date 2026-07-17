package connectwise

import (
	"maps"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	communicationTypeEmail = "Email"
	communicationTypePhone = "Phone"
	communicationTypeFax   = "Fax"
)

func attachCommunicationItems(node *ajson.Node, root map[string]any) error { // nolint:cyclop
	communicationItems, err := jsonquery.New(node).ArrayOptional("communicationItems")
	if err != nil {
		return err
	}

	if len(communicationItems) == 0 {
		// This contact doesn't have communication items.
		// Nothing to attach.
		return nil
	}

	fields := make(map[string]any)

	for _, commItem := range communicationItems {
		item, err := jsonquery.ParseNode[readCommunicationItem](commItem)
		if err != nil {
			return err
		}

		if !item.DefaultFlag {
			// Only the default communication items in each category matter.
			// Other items can be skipped as they don't participate in Read/Metadata.
			continue
		}

		switch item.CommunicationType {
		case communicationTypeEmail:
			fields[virtualFieldContactEmail] = item.Value
			fields[virtualFieldContactEmailId] = item.Type.Id.String()
		case communicationTypeFax:
			fields[virtualFieldContactFax] = item.Value
			fields[virtualFieldContactFaxId] = item.Type.Id.String()
		case communicationTypePhone:
			fields[virtualFieldContactPhone] = item.Value
			fields[virtualFieldContactPhoneId] = item.Type.Id.String()
		}
	}

	// Move communication items to the top root level.
	maps.Copy(root, fields)

	return nil
}

// readCommunicationItem is returned by contact as elements of `communicationItems` array.
type readCommunicationItem struct {
	Id   int `json:"id"`
	Type struct {
		Id   naming.Text `json:"id"`
		Name string      `json:"name"`
		Info any         `json:"_info"`
	} `json:"type"`
	Value             any    `json:"value"`
	DefaultFlag       bool   `json:"defaultFlag"`
	CommunicationType string `json:"communicationType"`
}
