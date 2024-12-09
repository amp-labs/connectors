package providers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrInvalidConnectorParam = errors.New("connector parameter missing or invalid")
	ErrMissingParam          = errors.New("missing param for initializing connector")
)

// Parameters refer to the input configuration you pass into a connector for initialization.
type Parameters map[ParameterKey]any
type ParameterKey string

var (
	// ParameterHTTPClient can be used to set the underlying HTTP client for the connector.
	ParameterHTTPClient ParameterKey = "httpClient"

	// ParameterAuthenticatedClient can be used to set the underlying client for the connector which knows
	// how to authenticate requests.
	ParameterAuthenticatedClient ParameterKey = "authenticatedClient"

	// ParameterModule can be used to indicate which particular provider API module to call.
	ParameterModule ParameterKey = "module"

	// ParameterMetadata can be used to pass in metadata that the connector needs per invocation.
	// For example, a workspace name, a region, etc. Generally these are strings that are used
	// to fill in the connector's provider info. If you need to replace a value inside the catalog,
	// put it in the metadata.
	ParameterMetadata ParameterKey = "metadata"
)

// ParamValues is a struct that holds the parsed values from the input configuration.
type ParamValues struct {
	HTTPClient          *http.Client
	AuthenticatedClient common.AuthenticatedHTTPClient
	Module              common.ModuleID
	Metadata            Metadata
}

// Metadata is a map of key-value pairs that can be used to pass in additional information to the connector.
type Metadata map[MetadataKey]string
type MetadataKey string

var (
	MetadataKeyWorkspace MetadataKey = "workspace"
	MetadataKeyServer    MetadataKey = "server"
	MetadataKeyCloudId   MetadataKey = "cloudId"
)

func ParseParams(opts Parameters, required ...ParameterKey) (*ParamValues, error) {
	// For required options, error out if any of them are missing
	for _, req := range required {
		_, err := MustGetConnectorParam[any](opts, req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrMissingParam, req)
		}
	}

	// Parse the options and extract values
	module := GetConnectorParam[*common.ModuleID](opts, ParameterModule, nil)
	authenticatedClient := GetConnectorParam[*http.Client](opts, ParameterAuthenticatedClient, nil)
	metadata := GetConnectorParam[Metadata](opts, ParameterMetadata, Metadata{})

	return &ParamValues{
		AuthenticatedClient: authenticatedClient,
		Module:              *module,
		Metadata:            metadata,
	}, nil
}

// GetConnectorParam returns the value of the given key, if present, safely cast to an assumed type.
// If the key is not present, or the value is not of the assumed type, it returns the
// zero value of the desired type, and false. In case of success, it returns the value and true.
func GetConnectorParam[A any](opts Parameters, key ParameterKey, defaultVal A) A { //nolint:ireturn
	if opts == nil {
		return defaultVal
	}

	val, present := opts[key]
	if !present {
		return defaultVal
	}

	a, ok := val.(A)
	if !ok {
		return defaultVal
	}

	return a
}

// MustGetConnectorParam retrieves a required parameter and returns an error if not present.
func MustGetConnectorParam[A any](opts Parameters, key ParameterKey) (A, error) {
	var zero A
	if opts == nil {
		return zero, fmt.Errorf("%w: %v", ErrInvalidConnectorParam, key)
	}

	val, present := opts[key]
	if !present {
		return zero, fmt.Errorf("%w: %v", ErrInvalidConnectorParam, key)
	}

	a, ok := val.(A)
	if !ok {
		return zero, fmt.Errorf("%w: %v", ErrInvalidConnectorParam, key)
	}
	return a, nil
}

func GetConnectorMetadata(metadata Metadata, key MetadataKey) string {
	return metadata[key]
}

func MustGetConnectorMetadata(metadata Metadata, key MetadataKey) (string, error) {
	val, present := metadata[key]
	if !present {
		return "", fmt.Errorf("%w: %v", ErrInvalidConnectorParam, key)
	}

	return val, nil
}
