package providers

// Catalog is the top-level structure of the configuration file.
type Catalog struct {
	Providers map[Provider]ProviderInfo `yaml:"providers"`
}

// ProviderInfo is the configuration for a specific provider.  We use reflection to substitute any variables
// in the configuration. The substitution is only done on string fields. If you want to use pointers in the struct,
// you might have to update the code to handle it.
type ProviderInfo struct {
	Support      ConnectorSupport  `validate:"required" yaml:"support"`
	AuthType     AuthType          `validate:"required" yaml:"authType"`
	BaseURL      string            `validate:"required" yaml:"baseUrl"`
	OauthOpts    OauthOpts         `yaml:"oauthOpts"`
	ProviderOpts map[string]string `yaml:"providerOptions,omitempty"`
}

type ConnectorSupport struct {
	Read      bool `yaml:"read"`
	Write     bool `yaml:"write"`
	BulkWrite bool `yaml:"bulkWrite"`
	Subscribe bool `yaml:"subscribe"`
	Proxy     bool `yaml:"proxy"`
}

type OauthOpts struct {
	AuthURL  string `yaml:"authUrl"`
	TokenURL string `yaml:"tokenUrl"`
}

type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
)

func (i *ProviderInfo) GetOption(key string) (string, bool) {
	if i.ProviderOpts == nil {
		return "", false
	}

	val, ok := i.ProviderOpts[key]

	return val, ok
}
