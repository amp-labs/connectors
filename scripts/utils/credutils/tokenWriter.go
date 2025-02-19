package credutils

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"gitlab.com/c0b/go-ordered-json"
	"golang.org/x/oauth2"
)

func WriteToken(defaultCredsFilePath string, token *oauth2.Token) error {
	credentials, err := internalCredsFileWrite(defaultCredsFilePath, token)
	if err != nil {
		return err
	}

	// Create/Update provider prefixed file.
	// Ex: keap-creds.json
	provider := credentials.Get(credscanning.Fields.Provider.PathJSON)

	providerName, ok := provider.(string)
	if !ok {
		return errors.New("cannot infer provider in credentials file") // nolint:err113
	}

	providerCredsFilePath := credscanning.LoadPath(providerName)
	_, err = internalCredsFileWrite(providerCredsFilePath, token)

	return err
}

func internalCredsFileWrite(filePath string, token *oauth2.Token) (*ordered.OrderedMap, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(jsonFile)

	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	orderedMap := ordered.NewOrderedMap()
	if err = orderedMap.UnmarshalJSON(fileData); err != nil {
		return nil, err
	}

	orderedMap.Set(credscanning.Fields.AccessToken.PathJSON, token.AccessToken)
	orderedMap.Set(credscanning.Fields.RefreshToken.PathJSON, token.RefreshToken)
	orderedMap.Set(credscanning.Fields.ExpiryFormat.PathJSON, "RFC3339Nano")
	orderedMap.Set(credscanning.Fields.Expiry.PathJSON, token.Expiry.UTC().Format(time.RFC3339Nano))

	outputData, err := json.MarshalIndent(orderedMap, "", "  ")
	if err != nil {
		return nil, err
	}

	return orderedMap, os.WriteFile(filePath, outputData, os.ModePerm) // nolint:gosec
}
