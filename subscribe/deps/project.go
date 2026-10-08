package deps

import "context"

// ProjectResolver resolves Ampersand project attributes.
type ProjectResolver interface {
	// GetProjectAppName returns the app name of the given Ampersand project.
	GetProjectAppName(ctx context.Context, projectID string) (string, error)
}
