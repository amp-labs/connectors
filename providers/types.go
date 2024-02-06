package providers

// Catalog is the top-level structure of the configuration file.
type Catalog struct {
	Providers map[Provider]ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Support ConnectorSupport `yaml:"support"`
	Auth    Auth             `yaml:"auth"`
	BaseURL string           `yaml:"baseUrl"`
}

type ConnectorSupport struct {
	Read      bool `yaml:"read"`
	Write     bool `yaml:"write"`
	BulkWrite bool `yaml:"bulkWrite"`
	Subscribe bool `yaml:"subscribe"`
	Proxy     bool `yaml:"proxy"`
}

type Auth struct {
	Type     AuthType `yaml:"type"`
	AuthURL  string   `yaml:"authUrl"`
	TokenURL string   `yaml:"tokenUrl"`
}

type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
)
