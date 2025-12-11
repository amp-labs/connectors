// nolint:revive,godoclint
package common

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ConnectorParams can be used to pass input parameters to the connector.
type ConnectorParams struct {
	Module ModuleID

	// AuthenticatedClient is an HTTP client for the connector that knows how to handle authentication for the provider.
	AuthenticatedClient AuthenticatedHTTPClient

	// Workspace is the provider workspace that the connector is connecting to. It could be a part of the metadata
	// map, but it is a common enough field that it is included here.
	Workspace string

	// Metadata is a map of key-value pairs that can be used to pass in additional information to the connector.
	// Generally this is used to substitute placeholders in the providerInfo, like workspace, server, etc, which is
	// information that is specific to the connection.
	Metadata map[string]string

	// CustomAuthenticatedClient [optional] is useful for connectors that work over non-http protocols or want
	// to use custom non-http clients. Connectors that need this will know to check for it & use it if available.
	CustomAuthenticatedClient any
}

var (
	ErrValidationFailed        = errors.New("validation failed")
	ErrMissingAuthClient       = errors.New("authenticated client not given")
	ErrMissingMetadata         = errors.New("metadata not given")
	ErrMissingWorkspace        = errors.New("workspace not given")
	ErrMissingCustomAuthClient = errors.New("custom authenticated client not given")
)

// ValidateParameters sees which interfaces conn implements, calls the relevant validation methods.
// nolint:cyclop
func ValidateParameters(conn any, params ConnectorParams) error {
	var errs error

	if r, ok := conn.(workspaceValidator); ok {
		errs = errors.Join(errs, r.validateWorkspace(params))
	}

	if r, ok := conn.(customAuthenticatedClientValidator); ok {
		errs = errors.Join(errs, r.validateCustomAuthenticatedClient(params))
	}

	if r, ok := conn.(authenticatedClientValidator); ok {
		errs = errors.Join(errs, r.validateAuthenticatedClient(params))
	}

	if r, ok := conn.(metadataValidator); ok {
		errs = errors.Join(errs, r.validateMetadata(params))
	}

	if r, ok := conn.(moduleValidator); ok {
		errs = errors.Join(errs, r.validateModule(params))
	}

	if errs != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, errs)
	}

	return nil
}

// customAuthenticatedClientValidator is an interface that requires a custom authenticated client
// to be set in the parameters.

type customAuthenticatedClientValidator interface {
	validateCustomAuthenticatedClient(parameters ConnectorParams) error
}

type RequireCustomAuthenticatedClient struct{}

func (RequireCustomAuthenticatedClient) validateCustomAuthenticatedClient(parameters ConnectorParams) error {
	if parameters.CustomAuthenticatedClient == nil {
		return ErrMissingCustomAuthClient
	}

	return nil
}

// authenticatedClientValidator is an interface that requires an authenticated client to be set in the parameters.

type authenticatedClientValidator interface {
	validateAuthenticatedClient(parameters ConnectorParams) error
}

type RequireAuthenticatedClient struct{}

func (RequireAuthenticatedClient) validateAuthenticatedClient(parameters ConnectorParams) error {
	if parameters.AuthenticatedClient == nil {
		return ErrMissingAuthClient
	}

	return nil
}

// workspaceValidator is an interface that requires a workspace to be set in the parameters.

type workspaceValidator interface {
	validateWorkspace(parameters ConnectorParams) error
}

type RequireWorkspace struct{}

func (RequireWorkspace) validateWorkspace(parameters ConnectorParams) error {
	if parameters.Workspace == "" {
		return ErrMissingWorkspace
	}

	return nil
}

// metadataValidator is an interface that requires metadata to be set in the parameters.

type metadataValidator interface {
	validateMetadata(parameters ConnectorParams) error
}

// RequireMetadata defines required metadata for validation.
// Metadata keys are case-insensitive.
type RequireMetadata struct {
	ExpectedMetadataKeys []string
}

func (m RequireMetadata) validateMetadata(parameters ConnectorParams) error {
	if parameters.Metadata == nil {
		return ErrMissingMetadata
	}

	metadataLower := make(map[string]string)
	for k, v := range parameters.Metadata {
		metadataLower[strings.ToLower(k)] = v
	}

	for _, key := range m.ExpectedMetadataKeys {
		lowerKey := strings.ToLower(key)
		if metadataLower[lowerKey] == "" {
			return fmt.Errorf("%w: expected key %s not found", ErrMissingMetadata, key)
		}
	}

	return nil
}

// moduleValidator is an interface that requires a module to be set in the parameters.

type moduleValidator interface {
	validateModule(parameters ConnectorParams) error
}

type RequireModule struct {
	ExpectedModules []ModuleID
}

func (r RequireModule) validateModule(parameters ConnectorParams) error {
	if parameters.Module == "" {
		return ErrMissingModule
	}

	if !slices.Contains(r.ExpectedModules, parameters.Module) {
		return ErrUnsupportedModule
	}

	return nil
}
