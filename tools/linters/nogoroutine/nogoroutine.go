package nogoroutine

import (
	"go/ast"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("nogoroutine", New)
}

// Settings for the nogoroutine linter.
type Settings struct {
	// ExcludePaths are file path patterns to exclude from checking (e.g., "internal/future", "internal/simultaneously")
	ExcludePaths []string `json:"exclude-paths"`
}

// NoGoroutine is the custom linter plugin.
type NoGoroutine struct {
	settings Settings
}

// New creates a new instance of the nogoroutine linter.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &NoGoroutine{settings: s}, nil
}

// BuildAnalyzers returns the analyzers for this linter.
func (n *NoGoroutine) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "nogoroutine",
			Doc:  "detects bare 'go' keyword usage - use future.Go() or simultaneously.Do() instead",
			Run:  n.run,
		},
	}, nil
}

// GetLoadMode returns the load mode for this linter.
func (n *NoGoroutine) GetLoadMode() string {
	return register.LoadModeSyntax
}

// run is the main analysis function that detects go statements.
func (n *NoGoroutine) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Check if this file should be excluded
		filename := pass.Fset.Position(file.Pos()).Filename
		if n.shouldExclude(filename) {
			continue
		}

		// Traverse the AST looking for GoStmt nodes
		ast.Inspect(file, func(node ast.Node) bool {
			if goStmt, ok := node.(*ast.GoStmt); ok {
				pass.Report(analysis.Diagnostic{
					Pos:     goStmt.Pos(),
					End:     goStmt.End(),
					Message: "Direct use of 'go' keyword is forbidden. Use future.Go() or simultaneously.Do() instead to ensure panic recovery and prevent unbounded goroutine spawning.",
				})
			}
			return true
		})
	}

	return nil, nil
}

// shouldExclude checks if a file path should be excluded from checking.
func (n *NoGoroutine) shouldExclude(filename string) bool {
	for _, pattern := range n.settings.ExcludePaths {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}
