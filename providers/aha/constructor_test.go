package aha

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestConstructor(t *testing.T) { // nolint:funlen,wsl
	t.Parallel()

	type pair[T any] struct {
		Option T
		Err    error
	}

	clientPool := []pair[common.AuthenticatedHTTPClient]{{
		Option: nil,
		Err:    common.ErrMissingAuthClient,
	}, {
		Option: http.DefaultClient,
		Err:    nil,
	}}

	modulesPool := []pair[common.ModuleID]{{
		Option: common.ModuleRoot,
		Err:    nil,
	}, {
		Option: "",
		Err:    nil,
	}, {
		Option: "unknown-module",
		Err:    nil,
	}}

	workspacePool := []pair[string]{{
		Option: "",
		Err:    common.ErrMissingWorkspace,
	}, {
		Option: "workspace-office",
		Err:    nil,
	}}

	type test struct {
		name        string
		params      common.Parameters
		expectedErr []error
	}

	tests := make([]test, 0)

	for i, client := range clientPool {
		for j, module := range modulesPool {
			for k, workspace := range workspacePool {
				tests = append(tests, test{
					name: fmt.Sprintf("Constructor client[%v] module[%v] workspace[%v]", i+1, j+1, k+1),
					params: common.Parameters{
						Module:              module.Option,
						AuthenticatedClient: client.Option,
						Workspace:           workspace.Option,
						Metadata:            nil,
					},
					expectedErr: datautils.SliceNoNil(client.Err, module.Err, workspace.Err),
				})
			}
		}
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connector, err := NewConnector(tt.params)
			_ = connector

			testutils.CheckErrors(t, tt.name, tt.expectedErr, err)
		})
	}
}
