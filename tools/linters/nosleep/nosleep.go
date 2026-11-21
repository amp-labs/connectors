package nosleep

import (
	"go/ast"
	"go/types"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

func init() {
	register.Plugin("nosleep", New)
}

// forbidden maps a resolved package path to the function names in it that must not be called.
var forbidden = map[string]map[string]bool{
	"time":                                 {"Sleep": true},
	"github.com/amp-labs/amp-common/utils": {"SleepCtx": true},
}

// Settings for the nosleep linter.
type Settings struct{}

// NoSleep is the custom linter plugin.
type NoSleep struct{}

// New creates a new instance of the nosleep linter.
func New(settings any) (register.LinterPlugin, error) {
	if _, err := register.DecodeSettings[Settings](settings); err != nil {
		return nil, err
	}

	return &NoSleep{}, nil
}

// BuildAnalyzers returns the analyzers for this linter.
func (n *NoSleep) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "nosleep",
			Doc:  "forbids time.Sleep and amp-common utils.SleepCtx in connector code - let the HTTP client retry and back off (DoBatch) instead",
			Run:  n.run,
		},
	}, nil
}

// GetLoadMode returns the load mode for this linter.
// Type information is needed to resolve the callee's package regardless of the import alias.
func (n *NoSleep) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func (n *NoSleep) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
			if !ok || fn.Pkg() == nil || !forbidden[fn.Pkg().Path()][fn.Name()] {
				return true
			}

			pass.Report(analysis.Diagnostic{
				Pos: call.Pos(),
				End: call.End(),
				Message: "Sleeping in connector code is forbidden (" + fn.Pkg().Name() + "." + fn.Name() + "). " +
					"Let the HTTP client retry and back off (DoBatch) instead of sleeping.",
			})

			return true
		})
	}

	return nil, nil
}
