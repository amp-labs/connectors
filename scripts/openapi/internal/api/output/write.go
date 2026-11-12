// nolint:forbidigo,godoclint
package output

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
)

var (
	ErrMissingFileExtension = errors.New("missing file extensions")
	ErrUnknownFileExtension = errors.New("unknown file extension")
)

func Write(fileName string, object any) error {
	slog.Info("writing output", "file", fileName)

	if strings.HasSuffix(fileName, ".json.gz") {
		return codec.WriteJsonZip(fileName, object)
	} else if strings.HasSuffix(fileName, ".json") {
		return codec.WriteJson(fileName, object)
	}

	extension := filepath.Ext(fileName)
	if extension == "" {
		return fmt.Errorf("%w: file name %v", ErrMissingFileExtension, fileName)
	}

	return fmt.Errorf("%w: %v", ErrUnknownFileExtension, extension)
}

func PrintUnusedRules(name string, rules filters.Rules) {
	unusedRules := rules.UnusedRules()
	if len(unusedRules) == 0 {
		// Nothing to print.
		return
	}

	fmt.Printf("[%v] Unused Path Rules:\n", name)

	for _, rule := range unusedRules {
		fmt.Printf("\t%v\n", rule)
	}
}

func PrintRedundantKeys(name string, registry core.TrackedMap[string, string]) {
	keys := registry.GetRedundantKeys()
	if len(keys) == 0 {
		// Nothing to print.
		return
	}

	fmt.Printf("[%v] Redundant Keys:\n", name)

	for _, key := range keys {
		fmt.Printf("\t%v\n", key)
	}
}

func PrintAppliedKeys(name string, registry core.TrackedMap[string, string]) {
	keys := registry.GetAppliedKeys()
	if len(keys) == 0 {
		// Nothing to print.
		return
	}

	fmt.Printf("[%v] Applied Keys:\n", name)

	for _, key := range keys {
		fmt.Printf("\t%v\n", key)
	}
}
