//nolint:gochecknoglobals
package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testAccessTokenReader = &JSONReader{
	FilePath: "testcred1.json",
	JSONPath: "$['accessToken']",
	CredKey:  "AccessToken",
}

var testPreset = []Reader{
	&EnvReader{
		EnvName: "TEST_ENV_CLIENT_ID",
		CredKey: "ClientId",
	},
	&JSONReader{
		FilePath: "testcred1.json",
		JSONPath: "$['useToken']",
		CredKey:  "UseToken",
	},
	&JSONReader{
		FilePath: "testcred2.json",
		JSONPath: "$['refreshToken']",
		CredKey:  "RefreshToken",
	},
	&JSONReader{
		FilePath: "testcred2.json",
		JSONPath: "$['providers'][0]['name']",
		CredKey:  "Provider",
	},
	&JSONReader{
		FilePath: "testcred2.json",
		JSONPath: "$['providers'][0]['number']",
		CredKey:  "ProviderNumber",
	},
}

func TestCredentialOptions(t *testing.T) {
	t.Parallel()

	if os.Setenv("TEST_ENV_CLIENT_ID", "clientId") != nil {
		t.Fatal("Error setting environment variable")
	}

	opts := NewCredentialsRegistry()

	require.NoError(t, opts.AddReader(testAccessTokenReader))

	require.NoError(t, opts.AddReaders(testPreset...))

	refreshToken, err := opts.GetString("RefreshToken")
	require.NoError(t, err)
	require.Equal(t, "refreshToken", refreshToken)

	accessToken, err := opts.GetString("AccessToken")
	require.NoError(t, err)
	require.Equal(t, "accessToken", accessToken)

	clientId, err := opts.GetString("ClientId")
	require.NoError(t, err)
	require.Equal(t, "clientId", clientId)

	provider, err := opts.GetString("Provider")
	require.NoError(t, err)
	require.Equal(t, "provider", provider)

	providerNumber, err := opts.GetFloat64("ProviderNumber")
	require.NoError(t, err)
	require.Equal(t, 3, int(providerNumber))

	useToken, err := opts.GetBool("UseToken")
	require.NoError(t, err)
	require.True(t, useToken)
}
