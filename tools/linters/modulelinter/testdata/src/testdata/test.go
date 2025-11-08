package testdata

import "github.com/amp-labs/connectors/common"

// Good example: Single module, no DefaultModule required
func goodExampleSingleModule() {
	SetInfo("provider1", ProviderInfo{
		Name:     "Provider One",
		AuthType: Oauth2,
		BaseURL:  "https://api.example.com",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Good example: No modules at all
func goodExampleNoModules() {
	SetInfo("provider2", ProviderInfo{
		Name:     "Provider Two",
		AuthType: Oauth2,
		BaseURL:  "https://api.example.com",
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Good example: Multiple modules WITH DefaultModule set
func goodExampleMultipleModulesWithDefault() {
	SetInfo("provider3", ProviderInfo{
		Name:          "Provider Three",
		AuthType:      Oauth2,
		BaseURL:       "https://api.example.com",
		DefaultModule: "module1",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
			"module2": ModuleInfo{
				BaseURL:     "https://api.example.com/v2",
				DisplayName: "Module 2",
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Bad example: Multiple modules WITHOUT DefaultModule
func badExampleMultipleModulesNoDefault() {
	SetInfo("provider4", ProviderInfo{ // want "ProviderInfo with multiple modules must have DefaultModule field set"
		Name:     "Provider Four",
		AuthType: Oauth2,
		BaseURL:  "https://api.example.com",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
			"module2": ModuleInfo{
				BaseURL:     "https://api.example.com/v2",
				DisplayName: "Module 2",
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Bad example: Multiple modules with empty string DefaultModule
func badExampleMultipleModulesEmptyDefault() {
	SetInfo("provider5", ProviderInfo{ // want "ProviderInfo with multiple modules must have DefaultModule field set"
		Name:          "Provider Five",
		AuthType:      Oauth2,
		BaseURL:       "https://api.example.com",
		DefaultModule: "",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
			"module2": ModuleInfo{
				BaseURL:     "https://api.example.com/v2",
				DisplayName: "Module 2",
			},
			"module3": ModuleInfo{
				BaseURL:     "https://api.example.com/v3",
				DisplayName: "Module 3",
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Dummy types to make the test compile
type Provider string
type AuthType string
type ModuleID = common.ModuleID

const Oauth2 AuthType = "oauth2"

type ProviderInfo struct {
	Name          string
	AuthType      AuthType
	BaseURL       string
	DefaultModule ModuleID
	Modules       *Modules
	Support       Support
}

type Modules map[ModuleID]ModuleInfo

type ModuleInfo struct {
	BaseURL     string
	DisplayName string
}

type Support struct {
	Proxy bool
	Read  bool
}

func SetInfo(provider Provider, info ProviderInfo) {
	// Dummy function for testing
}

// Good example: Provider with modules and ModuleDependencies properly set
func goodExampleWithModuleDeps() {
	SetInfo("provider7", ProviderInfo{
		Name:          "Provider Seven",
		AuthType:      Oauth2,
		BaseURL:       "https://api.example.com",
		DefaultModule: "module1",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Workspace ID",
					ModuleDependencies: &ModuleDependencies{
						"module1": ModuleDependency{},
					},
				},
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

// Bad example: Provider with modules but metadata input missing ModuleDependencies
func badExampleMissingModuleDeps() {
	SetInfo("provider8", ProviderInfo{
		Name:          "Provider Eight",
		AuthType:      Oauth2,
		BaseURL:       "https://api.example.com",
		DefaultModule: "module1",
		Modules: &Modules{
			"module1": ModuleInfo{
				BaseURL:     "https://api.example.com/v1",
				DisplayName: "Module 1",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{ // want "MetadataItemInput in multi-module provider must have ModuleDependencies set to non-nil value"
					Name:        "workspace",
					DisplayName: "Workspace ID",
				},
			},
		},
		Support: Support{
			Proxy: true,
			Read:  true,
		},
	})
}

type ProviderMetadata struct {
	Input []MetadataItemInput
}

type MetadataItemInput struct {
	Name               string
	DisplayName        string
	ModuleDependencies *ModuleDependencies
}

type ModuleDependencies map[ModuleID]ModuleDependency
type ModuleDependency map[string]interface{}
