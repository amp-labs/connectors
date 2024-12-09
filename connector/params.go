// nolint
package connector

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var ErrMissingValue = errors.New("missing metadata value")

// Parameters can be used to pass input parameters to the connector. The json tags help with validation. Do not
// remove these without updating the validation logic.
type Parameters struct {
	HTTPClient *http.Client
	Module     common.ModuleID

	// AuthenticatedClient is a client for the connector that knows how to handle authentication for the provider.
	AuthenticatedClient common.AuthenticatedHTTPClient

	// Workspace is the provider workspace that the connector is connecting to. It could be a part of the metadata
	// map, but it is a common enough field that it is included here.
	Workspace string

	// Metadata is a map of key-value pairs that can be used to pass in additional information to the connector.
	// Generally this is used to substitute placeholders in the providerInfo, like workspace, server, etc, which is
	// information that is specific to the connection.
	Metadata Metadata

	// Used by validation to determine if the parameters are valid.
	validity struct {
		invalid bool
		error   error
	}
}

type (
	Metadata    map[string]string
	MetadataKey string
)

var (
	MetadataKeyServer  MetadataKey = "server"
	MetadataKeyCloudId MetadataKey = "cloudId"
)

type Option func(*Parameters)

func mustAuthenticatedClient(params *Parameters) {
	if params.AuthenticatedClient == nil {
		params.validity.invalid = false
		params.validity.error = fmt.Errorf("%w: %s", ErrMissingValue, "AuthenticatedClient")
	}
}

func mustMetadata(params *Parameters) {
	if params.Metadata == nil {
		params.validity.invalid = false
		params.validity.error = fmt.Errorf("%w: %s", ErrMissingValue, "Metadata")
	}
}

func mustModule(params *Parameters) {
	if params.Module == "" {
		params.validity.invalid = false
		params.validity.error = fmt.Errorf("%w: %s", ErrMissingValue, "Module")
	}
}

func mustWorkspace(params *Parameters) {
	if params.Workspace == "" {
		params.validity.invalid = false
		params.validity.error = fmt.Errorf("%w: %s", ErrMissingValue, "Workspace")
	}
}

// MustGetConnectorMetadata returns the value of the key from the metadata map or an error.
func MustGetConnectorMetadata(metadata Metadata, key MetadataKey) (string, error) {
	val, present := metadata[string(key)]
	if !present {
		return "", fmt.Errorf("%w: %v", ErrMissingValue, key)
	}

	return val, nil
}
