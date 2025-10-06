//nolint:gochecknoglobals
package scanning

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testAccessTokenReader = &JSONReader{
	FilePath: "test/cred1.json",
	JSONPath: "$['accessToken']",
	KeyName:  "AccessToken",
}

var testPreset = []Reader{
	&EnvReader{
		EnvName: "TEST_ENV_CLIENT_ID",
		KeyName: "ClientId",
	},
	&JSONReader{
		FilePath: "test/cred1.json",
		JSONPath: "$['useToken']",
		KeyName:  "UseToken",
	},
	&JSONReader{
		FilePath: "test/cred2.json",
		JSONPath: "$['refreshToken']",
		KeyName:  "RefreshToken",
	},
	&JSONReader{
		FilePath: "test/cred2.json",
		JSONPath: "$['providers'][0]['name']",
		KeyName:  "Provider",
	},
	&JSONReader{
		FilePath: "test/cred2.json",
		JSONPath: "$['providers'][0]['number']",
		KeyName:  "ProviderNumber",
	},
	&ValueReader{
		Val:     "myValue",
		KeyName: "MyValue",
	},
}

func TestCredentialOptions(t *testing.T) {
	t.Parallel()

	if os.Setenv("TEST_ENV_CLIENT_ID", "clientId") != nil {
		t.Fatal("Error setting environment variable")
	}

	opts := NewRegistry()

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

	myValue, err := opts.GetString("MyValue")
	require.NoError(t, err)
	require.Equal(t, "myValue", myValue)

	require.Error(t,
		opts.AddReader(
			&ValueReader{},
		),
	)
	require.Error(t,
		opts.AddReaders(
			[]Reader{&ValueReader{}}...,
		),
	)
}
