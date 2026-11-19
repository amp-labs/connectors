package zendesksupport

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestConstructor(t *testing.T) { // nolint:funlen,wsl
	t.Parallel()

	type pair struct {
		Option Option
		Err    error
	}

	clientPool := []pair{{
		Option: nil,
		Err:    paramsbuilder.ErrMissingClient,
	}, {
		Option: WithAuthenticatedClient(nil),
		Err:    paramsbuilder.ErrMissingClient,
	}, {
		Option: WithAuthenticatedClient(http.DefaultClient),
		Err:    nil,
	}}

	modulesPool := []pair{{
		Option: nil,
		Err:    nil,
	}, {
		Option: WithModule(common.ModuleRoot),
		Err:    nil,
	}, {
		Option: WithModule(providers.ModuleZendeskTicketing),
		Err:    nil,
	}, {
		Option: WithModule(providers.ModuleZendeskHelpCenter),
		Err:    nil,
	}, {
		Option: WithModule(""),
		Err:    nil,
	}, {
		Option: WithModule("unknown-module"),
		Err:    nil,
	}}

	workspacePool := []pair{{
		Option: nil,
		Err:    paramsbuilder.ErrMissingWorkspace,
	}, {
		Option: WithWorkspace(""),
		Err:    paramsbuilder.ErrMissingWorkspace,
	}, {
		Option: WithWorkspace("workspace-office"),
		Err:    nil,
	}}

	type test struct {
		name        string
		options     []Option
		expectedErr []error
	}

	tests := make([]test, 0)

	for i, client := range clientPool {
		for j, module := range modulesPool {
			for k, workspace := range workspacePool {
				tests = append(tests, test{
					name:        fmt.Sprintf("Constructor client[%v] module[%v] workspace[%v]", i+1, j+1, k+1),
					options:     datautils.SliceNoNil(client.Option, module.Option, workspace.Option),
					expectedErr: datautils.SliceNoNil(client.Err, module.Err, workspace.Err),
				})
			}
		}
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connector, err := NewConnector(tt.options...)
			_ = connector

			testutils.CheckErrors(t, tt.name, tt.expectedErr, err)
		})
	}
}
