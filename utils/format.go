package utils

import "encoding/json"

func PrettyFormatStruct(s any) (string, error) {
	json, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}

	return string(json), nil
}
