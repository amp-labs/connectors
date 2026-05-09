// Command apex-trigger-gen generates an Apex trigger + handler class +
// companion test class as a Salesforce metadata-format directory ready for
// deployment via the Salesforce CLI (sf).
//
// Output structure (matches what providers/salesforce produces in-process):
//
//	<output-dir>/
//	├── package.xml
//	├── triggers/
//	│   ├── <TriggerName>.trigger
//	│   └── <TriggerName>.trigger-meta.xml
//	└── classes/
//	    ├── <TriggerName>_Handler.cls
//	    ├── <TriggerName>_Handler.cls-meta.xml
//	    ├── Test_<TriggerName>.cls
//	    └── Test_<TriggerName>.cls-meta.xml
//
// Usage:
//
//	go run ./scripts/salesforce/apex-trigger-gen \
//	  -object Account \
//	  -indicator-field amp_cdc_optimized__c \
//	  -watch-fields Name,Phone \
//	  -output-dir ./my-deploy
//
// Then deploy with the Salesforce CLI:
//
//	sf project deploy start \
//	  --metadata-dir ./my-deploy \
//	  --test-level RunSpecifiedTests \
//	  --tests Test_CDC_Account
//
// The "cdc" variant generates a trigger that flips a Boolean checkbox field
// to true when any watch field changes (used by the connector's Subscribe
// path for quota optimization). The "read" variant generates a trigger that
// stamps a Datetime field with System.now() on watch-field change (used by
// the filtered-read path).
package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
)

const (
	variantCDC  = "cdc"
	variantRead = "read"

	// dirPerm is the permission bits used for created output directories.
	dirPerm os.FileMode = 0o755

	// maxZipEntryBytes caps how much we extract from a single zip entry, as a
	// guard against decompression-bomb inputs even though we control the zip
	// producer here. 4 MiB is plenty for our generated Apex sources.
	maxZipEntryBytes = 4 << 20
)

var (
	errMissingObject         = errors.New("missing required flag: -object")
	errMissingIndicatorField = errors.New("missing required flag: -indicator-field")
	errMissingWatchFields    = errors.New("missing required flag: -watch-fields")
	errUnknownVariant        = errors.New(`unknown variant (expected "cdc" or "read")`)
	errZipPathEscape         = errors.New("zip entry path escapes output directory")
)

func main() {
	var (
		objectName     string
		triggerName    string
		variant        string
		indicatorField string
		watchFields    string
		outputDir      string
	)

	flag.StringVar(&objectName, "object", "",
		"Salesforce object the trigger runs on (e.g. Account). Required.")
	flag.StringVar(&triggerName, "trigger-name", "",
		"Override the generated trigger name. Default: 'CDC_<Object>' for cdc, 'Read_<Object>' for read.")
	flag.StringVar(&variant, "variant", variantCDC,
		`Trigger variant: "cdc" (Boolean indicator on watch-field change) or "read" (Datetime stamp).`)
	flag.StringVar(&indicatorField, "indicator-field", "",
		"API name of the indicator field the trigger maintains. Required.")
	flag.StringVar(&watchFields, "watch-fields", "",
		"Comma-separated list of field API names to watch for changes. Required.")
	flag.StringVar(&outputDir, "output-dir", "./apex-deploy",
		"Directory to write the metadata-format output to.")
	flag.Parse()

	if err := run(objectName, triggerName, variant, indicatorField, watchFields, outputDir); err != nil {
		log.Fatalf("apex-trigger-gen: %v", err)
	}
}

func run(objectName, triggerName, variant, indicatorField, watchFields, outputDir string) error {
	if objectName == "" {
		return errMissingObject
	}

	if indicatorField == "" {
		return errMissingIndicatorField
	}

	if watchFields == "" {
		return errMissingWatchFields
	}

	resolvedTriggerName, err := resolveTriggerName(triggerName, variant, objectName)
	if err != nil {
		return err
	}

	indicatorType, err := indicatorFieldTypeFor(variant)
	if err != nil {
		return err
	}

	params := salesforce.ApexTriggerParams{
		ObjectName:  objectName,
		TriggerName: resolvedTriggerName,
		IndicatorField: common.FieldDefinition{
			FieldName: indicatorField,
			ValueType: indicatorType,
		},
		WatchFields: splitWatchFields(watchFields),
	}

	zipData, err := constructTriggerZip(variant, params, indicatorField)
	if err != nil {
		return fmt.Errorf("failed to construct trigger zip: %w", err)
	}

	if err := extractZipToDir(zipData, outputDir); err != nil {
		return fmt.Errorf("failed to write metadata dir: %w", err)
	}

	// Test class name follows the documented "Test_<TriggerName>" convention
	// produced by the metadata generator.
	testClassName := "Test_" + resolvedTriggerName

	printDeployHint(outputDir, testClassName)

	return nil
}

func resolveTriggerName(override, variant, objectName string) (string, error) {
	if override != "" {
		return override, nil
	}

	switch variant {
	case variantCDC:
		return salesforce.GenerateApexTriggerNameForCDC(objectName)
	case variantRead:
		return salesforce.GenerateApexTriggerNameForRead(objectName)
	default:
		return "", fmt.Errorf("%w: got %q", errUnknownVariant, variant)
	}
}

func indicatorFieldTypeFor(variant string) (common.FieldType, error) {
	switch variant {
	case variantCDC:
		return common.FieldTypeBoolean, nil
	case variantRead:
		return common.FieldTypeDateTime, nil
	default:
		return "", fmt.Errorf("%w: got %q", errUnknownVariant, variant)
	}
}

// constructTriggerZip dispatches to the appropriate variant-specific wrapper
// in the salesforce package. The wrapper validates params (including
// indicatorField) and bundles the trigger + handler + test class into a zip.
func constructTriggerZip(variant string, params salesforce.ApexTriggerParams, indicatorField string) ([]byte, error) {
	switch variant {
	case variantCDC:
		return salesforce.ConstructApexTriggerZipForCDC(params, indicatorField)
	case variantRead:
		return salesforce.ConstructApexTriggerZipForFilteredRead(params, indicatorField)
	default:
		return nil, fmt.Errorf("%w: got %q", errUnknownVariant, variant)
	}
}

func splitWatchFields(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		out = append(out, p)
	}

	return out
}

// extractZipToDir writes every entry of the in-memory zip to disk under
// outputDir, preserving the relative path layout (triggers/..., classes/...,
// package.xml). Existing files in outputDir are overwritten.
func extractZipToDir(zipData []byte, outputDir string) error {
	if err := os.MkdirAll(outputDir, dirPerm); err != nil {
		return fmt.Errorf("failed to create output dir %q: %w", outputDir, err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to resolve output dir %q: %w", outputDir, err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read zip: %w", err)
	}

	for _, f := range reader.File {
		if err := writeZipEntry(f, absOutputDir); err != nil {
			return err
		}
	}

	return nil
}

func writeZipEntry(entry *zip.File, absOutputDir string) error {
	// Resolve the entry path relative to outputDir and reject any entry whose
	// resolved location escapes the directory (Zip-Slip / G305). The bare
	// filepath.Join would otherwise allow ../../etc/passwd-style escapes if a
	// hostile zip producer slipped one in.
	outPath := filepath.Join(absOutputDir, entry.Name) //nolint:gosec // G305: path validated below before any open/write
	if !strings.HasPrefix(outPath, absOutputDir+string(os.PathSeparator)) && outPath != absOutputDir {
		return fmt.Errorf("%w: %q", errZipPathEscape, entry.Name)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), dirPerm); err != nil {
		return fmt.Errorf("failed to create dir for %q: %w", outPath, err)
	}

	entryReader, err := entry.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry %q: %w", entry.Name, err)
	}
	defer entryReader.Close()

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create %q: %w", outPath, err)
	}
	defer out.Close()

	// io.CopyN with a hard cap guards against a malformed/oversized zip entry
	// (G110). io.EOF means the entry was smaller than the cap, which is the
	// expected case for our generated Apex sources.
	if _, err := io.CopyN(out, entryReader, maxZipEntryBytes); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to write %q: %w", outPath, err)
	}

	if _, err := fmt.Fprintf(os.Stdout, "wrote %s\n", outPath); err != nil {
		return fmt.Errorf("failed to print progress: %w", err)
	}

	return nil
}

func printDeployHint(outputDir, testClassName string) {
	fmt.Fprintf(os.Stdout, "\nGenerated metadata-format deploy at: %s\n\n", outputDir)
	fmt.Fprintln(os.Stdout, "Deploy with the Salesforce CLI:")
	fmt.Fprintf(os.Stdout, `
  sf project deploy start \
    --metadata-dir %s \
    --test-level RunSpecifiedTests \
    --tests %s

`, outputDir, testClassName)
}
