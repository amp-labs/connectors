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
	Metadata Metadata
}

type (
	Metadata    map[string]string
	MetadataKey string
)

var (
	ErrValidationFailed  = errors.New("validation failed")
	ErrMissingAuthClient = errors.New("authenticated client not given")
	ErrMissingMetadata   = errors.New("metadata not given")
	ErrMissingWorkspace  = errors.New("workspace not given")
)

// ValidateParameters sees which interfaces conn implements, calls the relevant validation methods.
// nolint:cyclop
func ValidateParameters(conn any, params Parameters) error {
	var errs []error

	if r, ok := conn.(requireWorkspace); ok {
		if err := r.validateWorkspace(params); err != nil {
			errs = append(errs, err)
		}
	}

	if r, ok := conn.(requireAuthenticatedClient); ok {
		if err := r.validateAuthenticatedClient(params); err != nil {
			errs = append(errs, err)
		}
	}

	if r, ok := conn.(requireMetadata); ok {
		if err := r.validateMetadata(params); err != nil {
			errs = append(errs, err)
		}
	}

	if r, ok := conn.(requireModule); ok {
		if err := r.validateModule(params); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		err := ErrValidationFailed

		for _, e := range errs {
			err = fmt.Errorf("%w: %w", err, e)
		}

		return err
	}

	return nil
}

// RequireAuthenticatedClient is an interface that requires an authenticated client to be set in the parameters.

type requireAuthenticatedClient interface {
	validateAuthenticatedClient(parameters Parameters) error
}

type RequireAuthenticatedClient struct{}

func (RequireAuthenticatedClient) validateAuthenticatedClient(parameters Parameters) error {
	if parameters.AuthenticatedClient == nil {
		return ErrMissingAuthClient
	}

	return nil
}

// RequireWorkspace is an interface that requires a workspace to be set in the parameters.

type requireWorkspace interface {
	validateWorkspace(parameters Parameters) error
}

type RequireWorkspace struct{}

func (RequireWorkspace) validateWorkspace(parameters Parameters) error {
	if parameters.Workspace == "" {
		return ErrMissingWorkspace
	}

	return nil
}

// RequireMetadata is an interface that requires metadata to be set in the parameters.

type requireMetadata interface {
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

// RequireModule is an interface that requires a module to be set in the parameters.

type requireModule interface {
	validateModule(parameters Parameters) error
}

type RequireModule struct{}

func (RequireModule) validateModule(parameters Parameters) error {
	if parameters.Module == "" {
		return ErrMissingModule
	}

	return nil
}
