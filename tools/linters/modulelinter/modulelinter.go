package modulelinter

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("modulelinter", New)
}

// Settings for the modulelinter linter.
type Settings struct {
	// No settings needed for this linter
}

// ModuleLinter is the custom linter plugin.
type ModuleLinter struct {
	settings Settings
}

// New creates a new instance of the modulelinter linter.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &ModuleLinter{settings: s}, nil
}

// BuildAnalyzers returns the analyzers for this linter.
func (m *ModuleLinter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "modulelinter",
			Doc:  "checks module-related rules in ProviderInfo: DefaultModule and ModuleDependencies",
			Run:  m.run,
		},
	}, nil
}

// GetLoadMode returns the load mode for this linter.
func (m *ModuleLinter) GetLoadMode() string {
	return register.LoadModeSyntax
}

// run is the main analysis function that detects ProviderInfo composite literals.
func (m *ModuleLinter) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Only check files in the providers package or testdata (for testing)
		if file.Name.Name != "providers" && file.Name.Name != "testdata" {
			continue
		}

		// Traverse the AST looking for SetInfo calls
		ast.Inspect(file, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if this is a SetInfo call (fast check)
			if !isSetInfoCall(callExpr) {
				return true
			}

			// SetInfo should have 2 arguments: provider name and ProviderInfo struct
			if len(callExpr.Args) != 2 {
				return true
			}

			// The second argument should be a composite literal (ProviderInfo{...})
			compositeLit, ok := callExpr.Args[1].(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Check if this is a ProviderInfo composite literal
			if !isProviderInfoLiteral(compositeLit) {
				return true
			}

			// Single pass through fields to find Modules, DefaultModule, and Metadata
			var modulesField, defaultModuleField, metadataField ast.Expr
			for _, elt := range compositeLit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				key, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}

				switch key.Name {
				case "Modules":
					modulesField = kv.Value
				case "DefaultModule":
					defaultModuleField = kv.Value
				case "Metadata":
					metadataField = kv.Value
				}

				// Early exit if we found all fields we care about
				if modulesField != nil && defaultModuleField != nil && metadataField != nil {
					break
				}
			}

			modulesMap := getModulesMap(modulesField)
			hasModules := modulesField != nil && len(modulesMap) > 0
			hasMultipleModules := hasModules && len(modulesMap) > 1

			// Rule 1: Check if multiple modules require DefaultModule
			if hasMultipleModules {
				if defaultModuleField == nil || isZeroValue(defaultModuleField) {
					pass.Report(analysis.Diagnostic{
						Pos:     compositeLit.Pos(),
						End:     compositeLit.End(),
						Message: "ProviderInfo with multiple modules must have DefaultModule field set",
					})
				}
			}

			// Rule 2: Check ModuleDependencies in metadata inputs when provider has modules
			if hasModules && metadataField != nil {
				m.checkMetadataModuleDependencies(pass, metadataField)
			}

			return true
		})
	}

	return nil, nil
}

// checkMetadataModuleDependencies validates that all MetadataItemInput have ModuleDependencies
// when the provider has modules.
func (m *ModuleLinter) checkMetadataModuleDependencies(pass *analysis.Pass, metadataField ast.Expr) {
	// metadataField is &ProviderMetadata{...}
	// We need to unwrap it and find the Input field

	// Handle unary expression (e.g., &ProviderMetadata{...})
	if unary, ok := metadataField.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		metadataField = unary.X
	}

	// Now it should be a composite literal
	metadataLit, ok := metadataField.(*ast.CompositeLit)
	if !ok {
		return
	}

	// Find the Input field in ProviderMetadata
	var inputField ast.Expr
	for _, elt := range metadataLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Input" {
			continue
		}

		inputField = kv.Value
		break
	}

	if inputField == nil {
		return // No Input field
	}

	// Input should be a composite literal of []MetadataItemInput
	inputLit, ok := inputField.(*ast.CompositeLit)
	if !ok {
		return
	}

	// Check each MetadataItemInput
	for _, item := range inputLit.Elts {
		itemLit, ok := item.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Check if this MetadataItemInput has ModuleDependencies field
		hasModuleDependencies := false
		for _, field := range itemLit.Elts {
			kv, ok := field.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "ModuleDependencies" {
				continue
			}

			// Found ModuleDependencies, check if it's non-nil
			if !isZeroValue(kv.Value) {
				hasModuleDependencies = true
			}
			break
		}

		if !hasModuleDependencies {
			pass.Report(analysis.Diagnostic{
				Pos:     itemLit.Pos(),
				End:     itemLit.End(),
				Message: "MetadataItemInput in multi-module provider must have ModuleDependencies set to non-nil value",
			})
		}
	}
}

// isSetInfoCall checks if the call expression is a SetInfo call.
func isSetInfoCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "SetInfo"
}

// isProviderInfoLiteral checks if the composite literal is a ProviderInfo.
func isProviderInfoLiteral(lit *ast.CompositeLit) bool {
	ident, ok := lit.Type.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "ProviderInfo"
}

// getModulesMap returns the map of modules from a Modules field expression.
// Returns empty map if not a valid modules map.
func getModulesMap(modulesValue ast.Expr) map[string]bool {
	result := make(map[string]bool)

	// Handle unary expression (e.g., &Modules{...})
	if unary, ok := modulesValue.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		modulesValue = unary.X
	}

	// Now it should be a composite literal
	compositeLit, ok := modulesValue.(*ast.CompositeLit)
	if !ok {
		return result
	}

	// Collect module IDs
	for _, elt := range compositeLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// The key can be either an identifier or a string literal
		switch key := kv.Key.(type) {
		case *ast.Ident:
			result[key.Name] = true
		case *ast.BasicLit:
			if key.Kind == token.STRING {
				// Remove quotes from string literal
				result[strings.Trim(key.Value, `"`)] = true
			}
		}
	}

	return result
}

// isZeroValue checks if the expression represents a zero value.
func isZeroValue(expr ast.Expr) bool {
	// Check for empty string ""
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Value == `""`
	}

	// Check for nil
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil"
	}

	return false
}
