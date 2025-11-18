package credscanning

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/iancoleman/strcase"
)

// Fields is a grouping of constant values that dictate the keys that
// can be present inside *creds.json file.
var Fields = struct { // nolint:gochecknoglobals
	Provider Field
	// Tokens
	AccessToken  Field
	RefreshToken Field
	Expiry       Field
	ExpiryFormat Field
	// Client ID, Secret
	ClientId     Field
	ClientSecret Field
	// Basic Authentication
	Username Field
	Password Field
	// Key
	ApiKey    Field
	ApiSecret Field
	// Catalog variables
	Workspace Field
	// Oauth2
	State  Field
	Scopes Field
	Secret Field
	Token  Field
}{
	Provider: Field{
		Name:      "provider",
		PathJSON:  "provider",
		SuffixENV: "PROVIDER",
	},
	AccessToken: Field{
		Name:      "accessToken",
		PathJSON:  "accessToken",
		SuffixENV: "ACCESS_TOKEN",
	},
	RefreshToken: Field{
		Name:      "refreshToken",
		PathJSON:  "refreshToken",
		SuffixENV: "REFRESH_TOKEN",
	},
	Expiry: Field{
		Name:      "expiry",
		PathJSON:  "expiry",
		SuffixENV: "EXPIRY",
	},
	ExpiryFormat: Field{
		Name:      "expiryFormat",
		PathJSON:  "expiryFormat",
		SuffixENV: "EXPIRY_FORMAT",
	},
	ClientId: Field{
		Name:      "clientId",
		PathJSON:  "clientId",
		SuffixENV: "CLIENT_ID",
	},
	ClientSecret: Field{
		Name:      "clientSecret",
		PathJSON:  "clientSecret",
		SuffixENV: "CLIENT_SECRET",
	},
	Username: Field{
		Name:      "username",
		PathJSON:  "username",
		SuffixENV: "USERNAME",
	},
	Password: Field{
		Name:      "password",
		PathJSON:  "password",
		SuffixENV: "PASSWORD",
	},
	ApiKey: Field{
		Name:      "apiKey",
		PathJSON:  "apiKey",
		SuffixENV: "API_KEY",
	},
	ApiSecret: Field{
		Name:      "apiSecret",
		PathJSON:  "apiSecret",
		SuffixENV: "API_SECRET",
	},
	Workspace: Field{
		Name:      "workspace",
		PathJSON:  "metadata.workspace",
		SuffixENV: "WORKSPACE",
	},
	State: Field{
		Name:      "state",
		PathJSON:  "state",
		SuffixENV: "STATE",
	},
	Scopes: Field{
		Name:      "scopes",
		PathJSON:  "scopes",
		SuffixENV: "SCOPES",
	},
	Secret: Field{
		Name:      "secret",
		PathJSON:  "secret",
		SuffixENV: "SECRET",
	},
	Token: Field{
		Name:      "token",
		PathJSON:  "token",
		SuffixENV: "TOKEN",
	},
}

type Field struct {
	Name      string
	PathJSON  string
	SuffixENV string
}

func (f Field) GetJSONReader(filepath string) *scanning.JSONReader {
	return &scanning.JSONReader{
		FilePath: filepath,
		JSONPath: jsonPathTo(f.PathJSON),
		KeyName:  f.Name,
	}
}

func (f Field) GetENVReader(providerName string) *scanning.EnvReader {
	return &scanning.EnvReader{
		EnvName: envNameFor(providerName, f.SuffixENV),
		KeyName: f.Name,
	}
}

// nolint:cyclop
func getFields(info providers.ProviderInfo,
	withRequiredAccessToken bool, customFields []Field,
) (datautils.NamedLists[Field], error) {
	lists := datautils.NamedLists[Field]{}
	requiredType := "required"
	optionalType := "optional"

	for _, customField := range customFields {
		lists.Add(optionalType, customField)
	}

	lists.Add(requiredType, Fields.Provider)

	if withRequiredAccessToken {
		lists.Add(requiredType, Fields.AccessToken)
		lists.Add(optionalType, Fields.RefreshToken, Fields.Expiry, Fields.ExpiryFormat)
	}

	switch info.AuthType {
	case providers.ApiKey:
		lists.Add(requiredType, Fields.ApiKey)
	case providers.Basic:
		lists.Add(requiredType, Fields.Username, Fields.Password)
	case providers.None:
	case providers.Oauth2:
		lists.Add(requiredType, Fields.ClientId, Fields.ClientSecret)
	case providers.Jwt:
		lists.Add(requiredType, Fields.Secret)
	case providers.Custom:
		// Custom auth may have different fields, so we skip adding any default fields here.
	default:
		return nil, ErrProviderInfo
	}

	var withRequiredWorkspace bool
	if info.Oauth2Opts != nil {
		// ExplicitWorkspaceRequired will be deprecated.
		withRequiredWorkspace = info.Oauth2Opts.ExplicitWorkspaceRequired
	}

	if info.RequiresWorkspace() || withRequiredWorkspace {
		lists.Add(requiredType, Fields.Workspace)
	}

	return lists, nil
}

func envNameFor(providerName string, suffix string) string {
	return fmt.Sprintf("%v_%v", envNameFormat(providerName), suffix)
}

func jsonPathTo(path string) string {
	return fmt.Sprintf("$.%v", path)
}

func envNameFormat(name string) string {
	name = strcase.ToSnake(name)

	return strings.ToUpper(name)
}
