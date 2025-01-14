package common

import (
	"errors"
	"fmt"
)

// Parameters can be used to pass input parameters to the connector.
type Parameters struct {
	Module ModuleID

	// AuthenticatedClient is a client for the connector that knows how to handle authentication for the provider.
	AuthenticatedClient AuthenticatedHTTPClient

	// Workspace is the provider workspace that the connector is connecting to. It could be a part of the metadata
	// map, but it is a common enough field that it is included here.
	Workspace string

	// Metadata is a map of key-value pairs that can be used to pass in additional information to the connector.
	// Generally this is used to substitute placeholders in the providerInfo, like workspace, server, etc, which is
	// information that is specific to the connection.
	Metadata map[string]string
}

var (
	ErrValidationFailed  = errors.New("validation failed")
	ErrMissingAuthClient = errors.New("authenticated client not given")
	ErrMissingMetadata   = errors.New("metadata not given")
	ErrMissingWorkspace  = errors.New("workspace not given")
)

// ValidateParameters sees which interfaces conn implements, calls the relevant validation methods.
// nolint:cyclop
func ValidateParameters(conn any, params Parameters) error {
	var errs error

	if r, ok := conn.(workspaceValidator); ok {
		errs = errors.Join(errs, r.validateWorkspace(params))
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

// authenticatedClientValidator is an interface that requires an authenticated client to be set in the parameters.

type authenticatedClientValidator interface {
	validateAuthenticatedClient(parameters Parameters) error
}

type RequireAuthenticatedClient struct{}

func (RequireAuthenticatedClient) validateAuthenticatedClient(parameters Parameters) error {
	if parameters.AuthenticatedClient == nil {
		return ErrMissingAuthClient
	}

	return nil
}

// workspaceValidator is an interface that requires a workspace to be set in the parameters.

type workspaceValidator interface {
	validateWorkspace(parameters Parameters) error
}

type RequireWorkspace struct{}

func (RequireWorkspace) validateWorkspace(parameters Parameters) error {
	if parameters.Workspace == "" {
		return ErrMissingWorkspace
	}

	return nil
}

// metadataValidator is an interface that requires metadata to be set in the parameters.

type metadataValidator interface {
	validateMetadata(parameters Parameters) error
}

type RequireMetadata struct {
	Expected []string
}

func (RequireMetadata) validateMetadata(parameters Parameters) error {
	if parameters.Metadata == nil {
		return ErrMissingMetadata
	}

	for _, key := range parameters.Metadata {
		if key == "" {
			return fmt.Errorf("%w: expected key %s not found", ErrMissingMetadata, key)
		}
	}

	return nil
}

// moduleValidator is an interface that requires a module to be set in the parameters.

type moduleValidator interface {
	validateModule(parameters Parameters) error
}

type RequireModule struct{}

func (RequireModule) validateModule(parameters Parameters) error {
	if parameters.Module == "" {
		return ErrMissingModule
	}

	return nil
}
