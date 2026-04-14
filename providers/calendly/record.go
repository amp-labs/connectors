package calendly

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// errEmptyResource is returned when a Calendly JSON envelope has no resource payload.
var errEmptyResource = errors.New("calendly: empty resource")

// GetRecordsByIds fetches event type resources by their canonical Calendly API URIs.
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	if objectName != "event_types" {
		return nil, fmt.Errorf("calendly: GetRecordsByIds unsupported for object %q: %w", objectName, common.ErrNotImplemented)
	}

	if len(recordIds) == 0 {
		return nil, common.ErrMissingRecordID
	}

	fieldSet := datautils.NewSetFromList(fields)
	out := make([]common.ReadResultRow, 0, len(recordIds))

	for _, recordURI := range recordIds {
		if recordURI == "" {
			continue
		}

		resp, err := c.JSONHTTPClient().Get(ctx, recordURI)
		if err != nil {
			return nil, fmt.Errorf("calendly: get %s: %w", recordURI, err)
		}

		parsed, err := common.UnmarshalJSON[struct {
			Resource map[string]any `json:"resource"`
		}](resp)
		if err != nil {
			return nil, err
		}

		if parsed.Resource == nil {
			return nil, fmt.Errorf("%w for %s", errEmptyResource, recordURI)
		}

		row := common.ReadResultRow{
			Fields: projectFields(parsed.Resource, fieldSet),
			Raw:    parsed.Resource,
			Id:     recordURI,
		}

		out = append(out, row)
	}

	return out, nil
}

func projectFields(resource map[string]any, fieldSet datautils.Set[string]) map[string]any {
	if len(fieldSet) == 0 {
		return lowercaseKeysCopy(resource)
	}

	out := make(map[string]any)

	for k, v := range resource {
		lk := strings.ToLower(k)
		if fieldSet.Has(lk) {
			out[lk] = v
		}
	}

	return out
}

func lowercaseKeysCopy(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))

	for k, v := range m {
		out[strings.ToLower(k)] = v
	}

	return out
}
