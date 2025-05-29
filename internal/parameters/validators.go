package parameters

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrValidationFailed  = errors.New("validation failed")
	ErrMissingAuthClient = errors.New("authenticated client not given")
	ErrMissingMetadata   = errors.New("metadata not given")
	ErrMissingWorkspace  = errors.New("workspace not given")
)

// Validate sees which interfaces connector implements, calls the relevant validation methods.
// nolint:cyclop
func (params Connector) Validate(conn any) error {
	var errs error

	if r, ok := conn.(moduleValidator); ok {
		errs = errors.Join(errs, r.validateModule(params))
	}

	if r, ok := conn.(workspaceValidator); ok {
		errs = errors.Join(errs, r.validateWorkspace(params))
	}

	if r, ok := conn.(authenticatedClientValidator); ok {
		errs = errors.Join(errs, r.validateAuthenticatedClient(params))
	}

	if r, ok := conn.(metadataValidator); ok {
		errs = errors.Join(errs, r.validateMetadata(params))
	}

	if errs != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, errs)
	}

	return nil
}

type moduleValidator interface {
	validateModule(parameters Connector) error
}

func (r RequireModule) validateModule(parameters Connector) error {
	if parameters.Module == "" {
		return common.ErrMissingModule
	}

	if !slices.Contains(r.ExpectedModules, parameters.Module) {
		return common.ErrUnsupportedModule
	}

	return nil
}

type authenticatedClientValidator interface {
	validateAuthenticatedClient(parameters Connector) error
}

func (RequireAuthenticatedClient) validateAuthenticatedClient(parameters Connector) error {
	if parameters.AuthenticatedClient == nil {
		return ErrMissingAuthClient
	}

	return nil
}

type workspaceValidator interface {
	validateWorkspace(parameters Connector) error
}

func (RequireWorkspace) validateWorkspace(parameters Connector) error {
	if parameters.Workspace == "" {
		return ErrMissingWorkspace
	}

	return nil
}

type metadataValidator interface {
	validateMetadata(parameters Connector) error
}

func (m RequireMetadata) validateMetadata(parameters Connector) error {
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
