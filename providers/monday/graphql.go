package monday

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type graphQLRequest struct {
	Query string `json:"query"`
}

func (c *Connector) postGraphQL(ctx context.Context, query string) (*common.JSONHTTPResponse, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	return c.JSONHTTPClient().Post(ctx, url.String(), graphQLRequest{Query: query})
}

func getItemsQuery(boardID string, limit int, cursor string, includeColumnValues bool) string {
	columnValuesFragment := ""
	if includeColumnValues {
		columnValuesFragment = `
			column_values {
				id
				text
				value
				type
			}`
	}

	cursorArg := ""
	if cursor != "" {
		cursorArg = fmt.Sprintf(`, cursor: "%s"`, escapeGraphQLString(cursor))
	}

	return fmt.Sprintf(`query {
		boards(ids: [%s]) {
			items_page(limit: %d%s) {
				cursor
				items {
					id
					name
					created_at
					updated_at
					board {
						id
					}
					group {
						id
						title
					}
					%s
				}
			}
		}
	}`, boardID, limit, cursorArg, columnValuesFragment)
}

func escapeGraphQLString(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func getCreateItemMutation(boardID, groupID, itemName, columnValuesJSON string) string {
	groupClause := ""
	if groupID != "" {
		groupClause = fmt.Sprintf(`group_id: "%s"`, escapeGraphQLString(groupID))
	}

	columnClause := ""
	if columnValuesJSON != "" {
		columnClause = fmt.Sprintf(`column_values: %s`, quoteGraphQLJSON(columnValuesJSON))
	}

	args := []string{
		fmt.Sprintf("board_id: %s", boardID),
		fmt.Sprintf(`item_name: "%s"`, escapeGraphQLString(itemName)),
	}
	if groupClause != "" {
		args = append(args, groupClause)
	}

	if columnClause != "" {
		args = append(args, columnClause)
	}

	return fmt.Sprintf(`mutation {
		create_item(%s) {
			id
			name
		}
	}`, strings.Join(args, ", "))
}

func getChangeMultipleColumnValuesMutation(boardID, itemID, columnValuesJSON string) string {
	return fmt.Sprintf(`mutation {
		change_multiple_column_values(
			board_id: %s,
			item_id: %s,
			column_values: %s
		) {
			id
		}
	}`, boardID, itemID, quoteGraphQLJSON(columnValuesJSON))
}

func getDeleteItemMutation(itemID string) string {
	return fmt.Sprintf(`mutation {
		delete_item(item_id: %s) {
			id
		}
	}`, itemID)
}

// quoteGraphQLJSON wraps a JSON object string as a GraphQL string literal argument.
func quoteGraphQLJSON(jsonPayload string) string {
	escaped := strings.ReplaceAll(jsonPayload, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)

	return fmt.Sprintf(`"%s"`, escaped)
}
