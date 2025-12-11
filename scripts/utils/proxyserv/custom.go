package proxyserv

import (
	"context"
	"fmt"
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
	httpClient := setupCustomHTTPClient(ctx, providerInfo, secrets, f.Debug, f.Metadata)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
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

	for _, input := range prov.CustomOpts.Inputs {
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
