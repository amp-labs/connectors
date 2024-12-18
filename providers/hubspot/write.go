package hubspot

import (
	"context"
	"fmt"
	"path"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	ErrInvalidDataFormat = fmt.Errorf("data must be a map")
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
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	relativeURL := path.Join("objects", config.ObjectName)
	url := c.getURL(relativeURL)

	if config.RecordId != "" {
		write = c.Client.Patch
		url = fmt.Sprintf("%s/%s", url, config.RecordId)
	} else {
		write = c.Client.Post
	}

	data, err := formatData(config.RecordData)
	if err != nil {
		return nil, err
	}

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

// formatData formats the data to be written to Hubspot. If the data contains a "properties" key, it's assumed to be
// formatted correctly. If not, it's wrapped in a "properties" object.
func formatData(data any) (map[string]any, error) {
	mapData, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrInvalidDataFormat, data)
	}

	// If the data has a "properties" / "associations" key, we assume it's formatted at the root level.
	if _, ok := mapData[string(ObjectFieldProperties)]; ok {
		return mapData, nil
	} else if _, ok := mapData[string(ObjectFieldAssociations)]; ok {
		return mapData, nil
	}

	return map[string]any{
		string(ObjectFieldProperties): data,
	}, nil
}
