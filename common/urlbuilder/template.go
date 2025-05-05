package urlbuilder

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

// Template is a URL factory that builds URLs by resolving named parts.
// It supports both constant and dynamic variables based on API requirements.
type Template struct {
	urlTemplate       string
	constantVariables []catalogreplacer.CatalogVariable
}

// NewTemplate creates new Template, which is a URL factory.
func NewTemplate(urlTemplate string, catalogVars ...catalogreplacer.CatalogVariable) *Template {
	return &Template{
		urlTemplate:       urlTemplate,
		constantVariables: catalogVars,
	}
}

// URL returns a new URL using only constant variables—no dynamic substitutions.
// Most connectors use simple URL templates where all parts are known ahead of time.
// If URL segments vary by request, use DynamicURL with dynamic variables instead.
func (t *Template) URL(path ...string) (*URL, error) {
	return t.DynamicURL(nil, path...)
}

// DynamicURL returns a new URL by applying dynamic variables to the template.
// Additional path segments can be appended, matching the New constructor.
func (t *Template) DynamicURL(dynamicVariables map[string]string, path ...string) (*URL, error) {
	catalogVars := paramsbuilder.NewCatalogVariables(dynamicVariables)
	catalogVars = append(catalogVars, t.constantVariables...)
	registry := catalogreplacer.NewCatalogSubstitutionRegistry(catalogVars)

	// Apply substitutions to the URL template.
	// The operation requires a struct, so the string is wrapped accordingly.
	template := &struct{ Data string }{
		Data: t.urlTemplate,
	}

	if err := registry.Apply(template); err != nil {
		return nil, err
	}

	return New(template.Data, path...)
}

// OverrideURL intended for enabling testing.
func (t *Template) OverrideURL(urlTemplate string) {
	t.urlTemplate = urlTemplate
}
