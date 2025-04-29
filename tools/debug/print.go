package debug

import "encoding/json"

//nolint:errchkjson
func PrettyFormatStringJSON(v any) string {
	prettyJSON, _ := json.MarshalIndent(v, "", "  ")

	return string(prettyJSON)
}
