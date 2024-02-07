package providers

// Catalog is the top-level structure of the configuration file.
type Catalog struct {
	Providers map[Provider]ProviderConfig `yaml:"providers"`
}

// ProviderConfig is the configuration for a specific provider.  We use reflection to substitute any variables
// in the configuration. The substitution is only done on string fields. If you want to use pointers in the struct,
// you might have to update the code to handle it.
type ProviderConfig struct {
	Support  ConnectorSupport `yaml:"support"`
	AuthType AuthType         `yaml:"authType"`
	AuthOpts AuthOpts         `yaml:"authOpts"`
	BaseURL  string           `yaml:"baseUrl"`
}

type ConnectorSupport struct {
	Read      bool `yaml:"read"`
	Write     bool `yaml:"write"`
	BulkWrite bool `yaml:"bulkWrite"`
	Subscribe bool `yaml:"subscribe"`
	Proxy     bool `yaml:"proxy"`
}

type AuthOpts struct {
	AuthURL  string `yaml:"authUrl"`
	TokenURL string `yaml:"tokenUrl"`
}

type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
)
