package connectwise

import (
	"context"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
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

		identifier := item.Type.Id.String()
		switch item.CommunicationType {
		case "Email":
			fields[virtualFieldContactEmail+identifier] = item.Value
			if item.DefaultFlag {
				fields[virtualFieldContactEmailDefault] = identifier
			}
		case "Fax":
			fields[virtualFieldContactFax+identifier] = item.Value
			if item.DefaultFlag {
				fields[virtualFieldContactFaxDefault] = identifier
			}
		case "Phone":
			fields[virtualFieldContactPhone+identifier] = item.Value
			if item.DefaultFlag {
				fields[virtualFieldContactPhoneDefault] = identifier
			}
		}
	}

	// Move communication items to the top root level.
	maps.Copy(root, fields)

	return nil
}

func (c *Connector) requestCommunicationTypes(ctx context.Context) ([]communicationTypeResponse, error) {
	typesUrl, err := c.getCommunicationTypesURL()
	if err != nil {
		return nil, err
	}

	result := make([]communicationTypeResponse, 0)
	url := typesUrl.String()

	// Paginated read until no `next` links is present in the header.
	for url != "" {
		res, err := c.JSONHTTPClient().Get(ctx, url, c.clientIdHeader())
		if err != nil {
			return nil, err
		}

		records, err := common.UnmarshalJSON[communicationTypesResponse](res)
		if err != nil {
			return nil, err
		}

		result = append(result, *records...)

		// Repeat for the next page if any.
		url = httpkit.HeaderLink(res, "next")
	}

	return result, nil
}

// communicationTypesResponse is returned by `/company/communicationTypes` endpoint.
type communicationTypesResponse []communicationTypeResponse

type communicationTypeResponse struct {
	Id          naming.Text `json:"id"`
	Description string      `json:"description"`
	PhoneFlag   bool        `json:"phoneFlag"`
	FaxFlag     bool        `json:"faxFlag"`
	EmailFlag   bool        `json:"emailFlag"`
	DefaultFlag bool        `json:"defaultFlag"`
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
