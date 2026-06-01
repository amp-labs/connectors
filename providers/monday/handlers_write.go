package monday

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) buildWriteMutation(
	ctx context.Context,
	params common.WriteParams,
	recordData map[string]any,
) (string, error) {
	switch params.ObjectName {
	case mondayObjectBoard:
		return buildBoardWriteMutation(params, recordData)
	case mondayObjectItems:
		return c.buildItemsWriteMutation(ctx, params, recordData)
	case mondayObjectUser:
		return "", ErrWriteUserNotSupported
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedObjectName, params.ObjectName)
	}
}

func buildBoardWriteMutation(params common.WriteParams, recordData map[string]any) (string, error) {
	if params.RecordId == "" {
		boardName, ok := recordData["name"].(string)
		if !ok {
			return "", ErrBoardNameRequired
		}

		return fmt.Sprintf(`mutation {
				create_board(board_name: "%s", board_kind: public) {
					id
					name
				}
			}`, boardName), nil
	}

	return fmt.Sprintf(`mutation {
				update_board(board_id: %s, board_attribute: name, new_value: "%v") {
					id
				}
			}`, params.RecordId, recordData["name"]), nil
}

func (c *Connector) buildItemsWriteMutation(
	ctx context.Context,
	params common.WriteParams,
	recordData map[string]any,
) (string, error) {
	boardID, err := boardIDFromWriteRecord(recordData)
	if err != nil {
		return "", err
	}

	columns, err := c.fetchBoardColumnDefinitions(ctx, boardID)
	if err != nil {
		return "", err
	}

	prepared, err := prepareItemWriteRecordData(recordData, columnDefinitionsByID(columns))
	if err != nil {
		return "", err
	}

	columnValuesJSON, _ := prepared["column_values"].(string)

	if params.RecordId == "" {
		itemName, ok := prepared["name"].(string)
		if !ok {
			return "", fmt.Errorf("%w: name is required for item creation", common.ErrMissingFields)
		}

		groupID, _ := prepared["group_id"].(string)

		return getCreateItemMutation(boardID, groupID, itemName, columnValuesJSON), nil
	}

	if columnValuesJSON == "" {
		return "", fmt.Errorf(
			"%w: column_values or cf_<columnId> fields required for item update",
			common.ErrMissingFields,
		)
	}

	return getChangeMultipleColumnValuesMutation(boardID, params.RecordId, columnValuesJSON), nil
}
