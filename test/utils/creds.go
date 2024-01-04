package utils

import (
	"os"

	"github.com/spyzhov/ajson"
)

// Credentials returns the credentials from the creds.json file.
func Credentials(defaultFileName string) (*ajson.Node, error) {
	fileName := defaultFileName

	if fn, ok := os.LookupEnv("CREDENTIALS_FILE"); ok {
		fileName = fn
	}

	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return ajson.Unmarshal(byteValue)
}
