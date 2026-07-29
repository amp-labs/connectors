package proxyserv

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"reflect"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/generic"
	"github.com/amp-labs/connectors/providers"
)

func (f Factory) CreateProxyCustom(ctx context.Context) *Proxy {
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	fields := getCustomFields(providerInfo)
	secrets := getCustomSecrets(f.Registry, fields)

	// Multi-step providers acquire their request-time secrets (referenced by the
	// header templates) by running the registered connect flow, the way the
	// Ampersand server would.
	if flow, ok := providers.CustomAuthFlowFor(providers.Provider(f.Provider)); ok {
		maps.Copy(secrets, f.runCustomAuthFlow(ctx, providerInfo, flow, secrets))
	}

	httpClient := setupCustomHTTPClient(ctx, providerInfo, secrets, f.Debug, f.Metadata)
	baseURL := f.getBaseURL()

	return newProxy(baseURL, httpClient)
}

// runCustomAuthFlow executes the HTTP connect steps of a multi-step custom auth
// flow and returns the acquired secrets. When the creds file supplies a
// "secrets" object, those values are returned instead and no HTTP calls are
// made — an offline mode for exercising the proxy without live credentials.
// Redirect steps need a browser and are not supported here.
func (f Factory) runCustomAuthFlow(
	ctx context.Context,
	prov *providers.ProviderInfo,
	flow providers.CustomAuthFlow,
	inputs map[string]string,
) map[string]string {
	if len(f.Secrets) > 0 {
		fmt.Fprintln(os.Stderr, "using pre-supplied secrets from creds file; skipping the connect flow")

		return f.Secrets
	}

	state := providers.NewAuthContext()

	for _, input := range prov.CustomOpts.Inputs {
		state.ConsumerInputs[input.Name] = inputs[input.Name]
	}

	for _, input := range prov.CustomOpts.ProviderInputs {
		state.ProviderInputs[input.Name] = inputs[input.Name]
	}

	maps.Copy(state.Metadata, f.Metadata)

	for i, step := range flow.ConnectSteps {
		if step.Redirect != nil {
			_, _ = fmt.Fprintf(os.Stderr, "connect step %d is a browser redirect; not supported by the proxy\n", i)
			os.Exit(1)
		}

		state = executeHTTPStep(ctx, state, *step.HTTP, i)
	}

	return state.Secrets
}

func executeHTTPStep(
	ctx context.Context, state providers.AuthContext, step providers.HTTPStep, index int,
) providers.AuthContext {
	state, req, err := step.BuildRequest(ctx, state)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "connect step %d: building request: %v\n", index, err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "connect step %d: %v\n", index, err)
		os.Exit(1)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		_, _ = fmt.Fprintf(os.Stderr, "connect step %d: %s %s returned %s: %s\n",
			index, req.Method, req.URL, resp.Status, string(body))
		os.Exit(1)
	}

	// ParseResponse handlers own closing the body.
	state, err = step.ParseResponse(ctx, state, resp)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "connect step %d: parsing response: %v\n", index, err)
		os.Exit(1)
	}

	return state
}

func forEachField(callback func(name string, f credscanning.Field)) {
	v := reflect.ValueOf(credscanning.Fields)
	t := v.Type()

	for i := range v.NumField() {
		name := t.Field(i).Name

		f, ok := v.Field(i).Interface().(credscanning.Field)
		if !ok {
			// If the field is not of type credscanning.Field, skip it
			continue
		}

		callback(name, f)
	}
}

func getCustomFields(prov *providers.ProviderInfo) []credscanning.Field {
	var fields []credscanning.Field

	var missing []string

	// ProviderInputs (builder-side keys like apiKey) are read from the same
	// creds file as consumer Inputs when testing locally.
	allInputs := append(
		append([]providers.CustomAuthInput{}, prov.CustomOpts.Inputs...),
		prov.CustomOpts.ProviderInputs...)

	for _, input := range allInputs {
		added := false

		forEachField(func(name string, f credscanning.Field) {
			if input.Name != f.Name {
				return
			}

			fields = append(fields, f)
			added = true
		})

		if !added {
			missing = append(missing, input.Name)
		}
	}

	if len(missing) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "missing fields in credscanning.Fields: %v\n", missing)
		os.Exit(1)
	}

	return fields
}

func setupCustomHTTPClient( //nolint:ireturn
	ctx context.Context,
	prov *providers.ProviderInfo,
	secretValues map[string]string,
	debug bool,
	metadata map[string]string,
) common.AuthenticatedHTTPClient {
	client, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: debug,
		CustomCreds: &providers.CustomAuthParams{
			Values: secretValues,
		},
	})
	if err != nil {
		panic(err)
	}

	cc, err := generic.NewConnector(prov.Name,
		generic.WithAuthenticatedClient(client),
		generic.WithMetadata(metadata),
	)
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

func getCustomSecrets(registry scanning.Registry, fields []credscanning.Field) map[string]string {
	secrets := make(map[string]string)

	for _, field := range fields {
		value := registry.MustString(field.Name)
		if value == "" {
			_, _ = fmt.Fprintln(os.Stderr, field.Name+" from registry is empty")
			os.Exit(1)
		}

		secrets[field.Name] = value
	}

	return secrets
}
