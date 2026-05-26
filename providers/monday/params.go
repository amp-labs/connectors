package monday

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
)

const (
	boardIDFilterKey = "board_id"
	// itemsObjectBoardSeparator joins object name and board id for ListObjectMetadata.
	// Example: items@1234567890
	itemsObjectBoardSeparator = "@"
)

var (
	// ErrBoardIDRequired is returned when board_id is required but missing.
	ErrBoardIDRequired = errors.New("board_id is required for items operations; set ReadParams.Filter to board_id=<id>, " +
		"BuilderFilter field board_id eq <id>, or ListObjectMetadata object name items@<board_id>")
)

// parseObjectNameAndBoardID splits items@<board_id> for metadata requests.
func parseObjectNameAndBoardID(objectName string) (string, string) {
	base, boardID, found := strings.Cut(objectName, itemsObjectBoardSeparator)
	if !found || boardID == "" {
		return objectName, ""
	}

	return base, boardID
}

// boardIDFromReadParams resolves board_id from structured or string filters.
func boardIDFromReadParams(params common.ReadParams) (string, error) {
	if params.BuilderFilter != nil {
		for _, filter := range params.BuilderFilter.FieldFilters {
			if strings.EqualFold(filter.FieldName, boardIDFilterKey) {
				if id, ok := filter.Value.(string); ok && id != "" {
					return id, nil
				}
			}
		}
	}

	if params.Filter != "" {
		if id := boardIDFromFilterString(params.Filter); id != "" {
			return id, nil
		}
	}

	return "", ErrBoardIDRequired
}

// boardIDFromFilterString parses board_id from Filter values such as "board_id=123".
func boardIDFromFilterString(filter string) string {
	for part := range strings.SplitSeq(filter, "&") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(key), boardIDFilterKey) {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// boardIDFromWriteRecord returns board_id from write payload or an error.
func boardIDFromWriteRecord(record map[string]any) (string, error) {
	if raw, ok := record[boardIDFilterKey]; ok {
		switch v := raw.(type) {
		case string:
			if v != "" {
				return v, nil
			}
		case float64:
			return fmt.Sprintf("%.0f", v), nil
		case int:
			return strconv.Itoa(v), nil
		case int64:
			return strconv.FormatInt(v, 10), nil
		}
	}

	return "", ErrBoardIDRequired
}
