package providers

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

var (
	testCatalog CatalogType = map[string]ProviderInfo{ // nolint:gochecknoglobals
		"test": {
			AuthType:    Oauth2,
			Name:        "test",
			BaseURL:     "https://{{.workspace}}.test.com",
			DisplayName: "Super Test",
		},
	}
	customTestCatalogOption = []CatalogOption{ // nolint:gochecknoglobals
		func(params *catalogParams) {
			params.catalog = &CatalogWrapper{
				Catalog:   testCatalog,
				Timestamp: time.Now().Format(time.RFC3339),
			}
		},
	}
)

func TestNewCustomCatalog(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name         string
		input        []CatalogOption
		expected     CatalogType
		expectedErrs []error
	}{
		{
			name: "Removing catalog is not allowed",
			input: []CatalogOption{
				func(params *catalogParams) {
					params.catalog = nil
				},
			},
			expected:     nil,
			expectedErrs: []error{ErrCatalogNotFound},
		},
		{
			name:         "Custom catalog can be set",
			input:        customTestCatalogOption,
			expected:     testCatalog,
			expectedErrs: nil,
		},
		{
			name:         "Builtin catalog is used by default",
			input:        []CatalogOption{},
			expected:     catalog,
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewCustomCatalog(tt.input...).catalog()
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if output != nil {
				if !reflect.DeepEqual(output.Catalog, tt.expected) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
				}
			}
		})
	}
}

func TestReadInfo(t *testing.T) { // nolint:funlen
	t.Parallel()

	type inType struct {
		options  []CatalogOption
		provider Provider
		vars     []catalogreplacer.CatalogVariable
	}

	tests := []struct {
		name         string
		input        inType
		expected     *ProviderInfo
		expectedErrs []error
	}{
		{
			name: "Returns missing provider error",
			input: inType{
				options:  customTestCatalogOption,
				provider: "nobody knows",
				vars:     nil,
			},
			expected:     nil,
			expectedErrs: []error{ErrProviderNotFound},
		},
		{
			name: "Works without substitution",
			input: inType{
				options:  customTestCatalogOption,
				provider: "test",
				vars:     nil,
			},
			expected: &ProviderInfo{
				AuthType:    Oauth2,
				Name:        "test",
				BaseURL:     "https://{{.workspace}}.test.com",
				DisplayName: "Super Test",
			},
			expectedErrs: nil,
		},
		{
			name: "Works with substitution",
			input: inType{
				options:  customTestCatalogOption,
				provider: "test",
				vars:     createCatalogVars("workspace", "europe"),
			},
			expected: &ProviderInfo{
				AuthType:    Oauth2,
				Name:        "test",
				BaseURL:     "https://europe.test.com",
				DisplayName: "Super Test",
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewCustomCatalog(tt.input.options...).
				ReadInfo(tt.input.provider, tt.input.vars...)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

func TestReadModuleInfo(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	type inType struct {
		provider Provider
		vars     []catalogreplacer.CatalogVariable
		moduleID common.ModuleID
	}

	tests := []struct {
		name     string
		input    inType
		expected *ModuleInfo
		// TODO this method should check: `expectedErr error`
	}{
		// Root for providers that have no modules.
		{
			name: "Dynamics root module",
			input: inType{
				provider: DynamicsCRM,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: common.ModuleRoot,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.api.crm.dynamics.com/api/data",
				DisplayName: "Microsoft Dynamics CRM",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Capsule root module",
			input: inType{
				provider: Capsule,
				moduleID: common.ModuleRoot,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.capsulecrm.com/api",
				DisplayName: "Capsule",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		// Root for providers that have multiple modules.
		{
			name: "Hubspot root module",
			input: inType{
				provider: Hubspot,
				moduleID: common.ModuleRoot,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.hubapi.com",
				DisplayName: "HubSpot",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Marketo root module",
			input: inType{
				provider: Marketo,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: common.ModuleRoot,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.mktorest.com",
				DisplayName: "Marketo",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Zoom root module",
			input: inType{
				provider: Zoom,
				moduleID: common.ModuleRoot,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.zoom.us",
				DisplayName: "Zoom",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		// Unknown module for providers with no modules.
		{
			name: "Dynamics unknown module",
			input: inType{
				provider: DynamicsCRM,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: "random-module-name",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.api.crm.dynamics.com/api/data",
				DisplayName: "Microsoft Dynamics CRM",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
			// expectedErr: common.ErrMissingModule,
		},
		{
			name: "Capsule unknown module",
			input: inType{
				provider: Capsule,
				moduleID: "random-module-name",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.capsulecrm.com/api",
				DisplayName: "Capsule",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
			// expectedErr: common.ErrMissingModule,
		},
		// Unknown module for providers with multiple modules fallbacks to default.
		{
			name: "Atlassian unknown module",
			input: inType{
				provider: Atlassian,
				moduleID: "random-module-name",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api",
				DisplayName: "Atlassian Jira",
				Support: Support{
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Hubspot unknown module",
			input: inType{
				provider: Hubspot,
				moduleID: "random-module-name",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.hubapi.com/crm",
				DisplayName: "HubSpot CRM",
				Support: Support{
					BatchWrite: &BatchWriteSupport{
						Create: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
						Update: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
					},
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Marketo unknown module",
			input: inType{
				provider: Marketo,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: "random-module-name",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.mktorest.com",
				DisplayName: "Marketo",
				Support: Support{
					BulkWrite: BulkWriteSupport{
						Insert: false,
						Update: false,
						Upsert: false,
						Delete: false,
					},
					Proxy:     true,
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		// Choosing non-root module for providers supporting several modules.
		{
			name: "Atlassian Jira module",
			input: inType{
				provider: Atlassian,
				moduleID: ModuleAtlassianJira,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api",
				DisplayName: "Atlassian Jira",
				Support: Support{
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Hubspot CRM module",
			input: inType{
				provider: Hubspot,
				moduleID: ModuleHubspotCRM,
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.hubapi.com/crm",
				DisplayName: "HubSpot CRM",
				Support: Support{
					BatchWrite: &BatchWriteSupport{
						Create: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
						Update: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
					},
					Read:  true,
					Write: true,
				},
			},
		},
		// Empty module for providers that have no modules defaults to root.
		{
			name: "Dynamics empty module",
			input: inType{
				provider: DynamicsCRM,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: "",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.api.crm.dynamics.com/api/data",
				DisplayName: "Microsoft Dynamics CRM",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		{
			name: "Capsule empty module",
			input: inType{
				provider: Capsule,
				moduleID: "",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://api.capsulecrm.com/api",
				DisplayName: "Capsule",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
		},
		// Choosing empty module for providers supporting several modules uses default from the Catalog.
		{
			name: "Marketo fallback to default module",
			input: inType{
				provider: Marketo,
				vars:     createCatalogVars("workspace", "london"),
				moduleID: "",
			},
			expected: &ModuleInfo{
				BaseURL:     "https://london.mktorest.com",
				DisplayName: "Marketo",
				Support: Support{
					BulkWrite: BulkWriteSupport{
						Insert: false,
						Update: false,
						Upsert: false,
						Delete: false,
					},
					Proxy:     true,
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := NewCustomCatalog().ReadInfo(tt.input.provider, tt.input.vars...)
			if err != nil {
				t.Fatalf("%s: bad test, failed to read info: (%v)", tt.name, err)
			}

			output := info.ReadModuleInfo(tt.input.moduleID)
			testutils.CheckOutput(t, tt.name, tt.expected, output)
		})
	}
}

func createCatalogVars(pairs ...string) []catalogreplacer.CatalogVariable {
	if len(pairs)%2 != 0 {
		return nil
	}

	result := make([]catalogreplacer.CatalogVariable, 0, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		j := i + 1

		result = append(result, catalogreplacer.CustomCatalogVariable{
			Plan: catalogreplacer.SubstitutionPlan{
				From: pairs[i],
				To:   pairs[j],
			},
		})
	}

	return result
}
