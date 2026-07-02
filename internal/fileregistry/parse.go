package fileregistry

import "encoding/json"

func MustParseJSON[D any](fileData []byte) D {
	var data D

	if err := json.Unmarshal(fileData, &data); err != nil {
		return data
	}

	return data
}
