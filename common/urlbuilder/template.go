package urlbuilder

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

// Template is a URL factory that builds URLs by resolving named parts.
// It supports both constant and dynamic variables based on API requirements.
type Template struct {
	Format            string
	constantVariables []catalogreplacer.CatalogVariable
}

// NewTemplate creates new Template, which is a URL factory.
func NewTemplate(urlFormat string, catalogVars ...catalogreplacer.CatalogVariable) *Template {
	return &Template{
		Format:            urlFormat,
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
		Data: t.Format,
	}

	if err := registry.Apply(template); err != nil {
		return nil, err
	}

	return New(template.Data, path...)
}

// RawURL is like URL but joins additional path segments as-is, without url encoding of path.
func (t *Template) RawURL(path ...string) (string, error) {
	return t.RawDynamicURL(nil, path...)
}

// RawDynamicURL behaves like DynamicURL but preserves the raw path segments.
// The path is treated as a list of opaque strings.
// No URL encoding or transformation is performed. Only proper slash separation is ensured.
func (t *Template) RawDynamicURL(dynamicVariables map[string]string, path ...string) (string, error) {
	url, err := t.DynamicURL(dynamicVariables)
	if err != nil {
		return "", err
	}

	return joinURL(url.String(), path...), nil
}
