package providers

import (
	"errors"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const (
	// configFileLoc is the name of the config file.
	configFileLoc = "providers.yaml"
)

var (
	ErrProviderConfigNotFound = errors.New("provider config not found")
)

// Config is the entire configuration for all providers.
type Config struct {
	Providers map[Provider]map[string]string `yaml:"providers"`
}

// ReadConfig reads the configuration from the config file for
// a specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}.
// The provider YAML has more details on how it works.
func ReadConfig(provider Provider, substitutions map[string]string) (map[string]string, error) {
	config, err := read()
	if err != nil {
		return nil, err
	}

	providerConfig, ok := config.Providers[provider]
	if !ok {
		return nil, ErrProviderConfigNotFound
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	for providerConfigKey, providerConfigValue := range providerConfig {
		providerConfig[providerConfigKey], err = substitute(providerConfigValue, &substitutions)
		if err != nil {
			return nil, err
		}
	}

	return providerConfig, nil
}

// substitute performs string substitution on the input string
// using the substitutions map.
func substitute(input string, substitutions *map[string]string) (string, error) {
	tmpl, err := template.New("-").Parse(input)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, substitutions)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// read reads the entire configuration from the config file.
func read() (*Config, error) {
	data, err := os.ReadFile(configFileLoc)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
