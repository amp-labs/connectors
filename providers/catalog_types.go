package providers

// ================================================================================
// Provider list
// ================================================================================

// Provider is the name of a provider.
type Provider string

// List all providers here.
const (
	Salesforce Provider = "salesforce"
	Hubspot    Provider = "hubspot"
	LinkedIn   Provider = "linkedIn"
)

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}

// ================================================================================
// Provider catalog structure
// ================================================================================

// CatalogType is the top-level structure of the configuration file.
type CatalogType map[Provider]ProviderInfo

// ProviderInfo is the configuration for a specific provider.  We use reflection to substitute any variables
// in the configuration. The substitution is only done on string fields. If you want to use pointers in the struct,
// you might have to update the substitution function to handle it.
type ProviderInfo struct {
	AuthType     AuthType `validate:"required"`
	Support      Support  `validate:"required"`
	BaseURL      string   `validate:"required"`
	OauthOpts    OauthOpts
	ProviderOpts map[string]string
}

type AuthType string

const AuthTypeOAuth2 AuthType = "oauth2"

type OauthOpts struct {
	AuthURL  string
	TokenURL string
}

type Support struct {
	Read      bool
	Write     bool
	BulkWrite bool
	Subscribe bool
	Proxy     bool
}

func (i *ProviderInfo) GetOption(key string) (string, bool) {
	if i.ProviderOpts == nil {
		return "", false
	}

	val, ok := i.ProviderOpts[key]

	return val, ok
}
