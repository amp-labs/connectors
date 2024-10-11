package deep

import "github.com/amp-labs/connectors/internal/deep/requirements"

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
