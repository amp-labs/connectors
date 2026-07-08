package utils

import (
	"encoding/json"
	"io"

	"github.com/amp-labs/connectors/common"
)

func PrintReadResultWithoutRaw(result *common.ReadResult, w io.Writer) {
	convertedValue := substituteErrorsToStrings(result)

	type readResultRow struct {
		Fields       map[string]any              `json:"fields"`
		Associations map[string][]map[string]any `json:"associations,omitempty"`
		Id           string                      `json:"id,omitempty"`
	}

	type readResult struct {
		Rows     int64           `json:"rows"`
		Data     []readResultRow `json:"data"`
		NextPage string          `json:"nextPage,omitempty"`
		Done     bool            `json:"done,omitempty"`
	}

	data, err := json.Marshal(convertedValue)
	if err != nil {
		Fail("error marshaling convertedValue: %w", "error", err)
	}

	readData := readResult{}
	if err = json.Unmarshal(data, &readData); err != nil {
		Fail("error unmarshaling data into readData: %w", "error", err)
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err = encoder.Encode(readData); err != nil {
		Fail("error marshaling to JSON: %w", "error", err)
	}
}
