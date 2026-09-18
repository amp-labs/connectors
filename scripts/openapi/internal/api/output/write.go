// nolint:forbidigo,godoclint
package output

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
)

var (
	ErrMissingFileExtension = errors.New("missing file extensions")
	ErrUnknownFileExtension = errors.New("unknown file extension")
)

func Write(fileName string, object any) error {
	slog.Info("writing output", "file", fileName)

	_, extension, ok := strings.Cut(fileName, ".")
	if !ok {
		return fmt.Errorf("%w: file name %v", ErrMissingFileExtension, fileName)
	}

	switch extension {
	case "json":
		return writeJson(fileName, object)
	case "json.gz":
		return writeJsonZip(fileName, object)
	default:
		return fmt.Errorf("%w: %v", ErrUnknownFileExtension, extension)
	}
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

func writeJson(filename string, object any) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(filename, data, os.ModePerm) //nolint:gosec
}

func writeJsonZip(filename string, object any) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return err
	}

	return file.Sync()
}
