package hubspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/internal/datautils"
)

type writeResponse struct {
	CreatedAt             string         `json:"createdAt"`
	Archived              bool           `json:"archived"`
	ArchivedAt            string         `json:"archivedAt"`
	PropertiesWithHistory any            `json:"propertiesWithHistory"`
	ID                    string         `json:"id"`
	Properties            map[string]any `json:"properties"`
	UpdatedAt             string         `json:"updatedAt"`
}

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	relativeURL := strings.Join([]string{"objects", config.ObjectName}, "/")

	url, err := c.getURL(relativeURL)
	if err != nil {
		return nil, err
	}

	if config.RecordId != "" {
		write = c.Client.Patch
		url = fmt.Sprintf("%s/%s", url, config.RecordId)
	} else {
		write = c.Client.Post
	}

	// Hubspot requires everything to be wrapped in a "properties" object.
	// We do this automatically in the write method so that the user doesn't
	// have to worry about it.
	data := make(map[string]any)
	data["properties"] = config.RecordData
	data["associations"] = config.Associations

	json, err := write(ctx, url, data)
	if err != nil {
		return nil, err
	}

	rsp, err := common.UnmarshalJSON[writeResponse](json)
	if err != nil {
		return nil, err
	}

	record, err := datautils.StructToMap(*rsp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		RecordId: rsp.ID,
		Success:  true,
		Data:     record,
	}, nil
}
