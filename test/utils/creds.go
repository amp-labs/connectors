package utils

import (
	"io"
	"os"

	"github.com/spyzhov/ajson"
)

func Credentials() (*ajson.Node, error) {
	fileName := "creds.json"

	if fn, ok := os.LookupEnv("CREDENTIALS_FILE"); ok {
		fileName = fn
	}

	cred, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = cred.Close()
	}()

	byteValue, err := io.ReadAll(cred)
	if err != nil {
		return nil, err
	}

	return ajson.Unmarshal(byteValue)
}
