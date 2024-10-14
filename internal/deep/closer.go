package deep

import "github.com/amp-labs/connectors/internal/deep/requirements"

// EmptyCloser is a major connector component which provides Close functionality.
// Embed this into connector struct.
// It is a no-op closer.
type EmptyCloser struct{}

func (EmptyCloser) Close() error {
	return nil
}

func (c EmptyCloser) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID: requirements.Closer,
		Constructor: func() *EmptyCloser {
			return &EmptyCloser{}
		},
	}
}
