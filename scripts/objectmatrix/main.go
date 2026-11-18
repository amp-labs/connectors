// Command objectmatrix generates the provider-by-object support matrix.
//
// It reads two sources and writes one artifact:
//
//   - each provider's shipped schemas.json, which is generated from the provider's own API
//     and is therefore the grounded answer for providers that have a fixed object list;
//   - the hand-declared entries in providers/objects/declared.go, for the providers that
//     operate on whatever object the credential can reach and so have no fixed list.
//
// Per-object operation support is taken from the provider's module Support flags. That is
// the finest granularity the catalog carries: a provider that declares write support is
// recorded as supporting write on each of its objects. Where a provider supports an
// operation for some objects and not others, this over-reports, which is why every entry
// carries the source it came from.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/objects"
)

const writePerm = 0o644

// schemaFile is the shape of a provider's shipped schemas.json.
type schemaFile struct {
	Modules map[string]struct {
		Objects map[string]json.RawMessage `json:"objects"`
	} `json:"modules"`
}

func main() {
	catalog, err := providers.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

	matrix := objects.Matrix{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Providers: make(map[string]objects.ProviderObjects),
	}

	schemaIndex := buildSchemaIndex()

	for name, info := range catalog.Catalog {
		entry := buildProviderEntry(name, info, schemaIndex[strings.ToLower(name)])
		if len(entry.Objects) > 0 || entry.Generic {
			matrix.Providers[name] = entry
		}
	}

	out, err := marshalWithoutEscaping(matrix)
	if err != nil {
		log.Fatal(err)
	}

	target := filepath.Join("providers", "objects", "matrix.json")

	err = os.WriteFile(target, out, writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Wrote %d providers to %s\n", len(matrix.Providers), target)
}

func buildProviderEntry(
	name string, info providers.ProviderInfo, schemaObjects map[string][]string,
) objects.ProviderObjects {
	entry := objects.ProviderObjects{
		Provider: name,
		Generic:  objects.IsGeneric(name),
		Objects:  make(map[string]map[string]objects.ObjectSupport),
	}

	// Objects the provider ships a schema for.
	for moduleID, objectNames := range schemaObjects {
		support := supportForModule(info, moduleID)
		addObjects(entry, moduleID, objectNames, support, objects.SourceSchema)
	}

	// Objects declared by hand, for providers with no fixed list to read.
	for moduleID, objectNames := range objects.DeclaredObjects(name) {
		support := supportForModule(info, moduleID)
		addObjects(entry, moduleID, objectNames, support, objects.SourceDeclared)
	}

	return entry
}

// addObjects records one module's objects, leaving any entry that is already present --
// a shipped schema is generated from the provider's API and so outranks a hand-written
// declaration for the same object.
func addObjects(
	entry objects.ProviderObjects,
	moduleID string,
	objectNames []string,
	support objects.ObjectSupport,
	source objects.Source,
) {
	if _, ok := entry.Objects[moduleID]; !ok {
		entry.Objects[moduleID] = make(map[string]objects.ObjectSupport)
	}

	for _, objectName := range objectNames {
		if _, exists := entry.Objects[moduleID][objectName]; exists {
			continue
		}

		withSource := support
		withSource.Source = source
		entry.Objects[moduleID][objectName] = withSource
	}
}

// supportForModule resolves the operations a module supports, falling back to the
// provider's top-level support when the module declares none.
func supportForModule(info providers.ProviderInfo, moduleID string) objects.ObjectSupport {
	support := info.Support

	if info.Modules != nil && moduleID != "" {
		if module, ok := (*info.Modules)[moduleID]; ok {
			support = module.Support
		}
	}

	return objects.ObjectSupport{
		Read:      support.Read,
		Write:     support.Write,
		Delete:    support.Delete,
		Subscribe: support.Subscribe,
	}
}

// schemaIndex maps a lowercased provider directory name to the objects its shipped
// schemas.json declares, keyed by module.
//
// Built by walking providers/ rather than by looking for providers/<name>/schemas.json:
// most of the 78 schema files sit a level or two deeper (providers/gusto/metadata/...),
// and a handful of directory names differ in case from the catalog name.
func buildSchemaIndex() map[string]map[string][]string {
	index := make(map[string]map[string][]string)

	err := filepath.WalkDir("providers", func(path string, entry fs.DirEntry, walkErr error) error {
		// A problem reading one entry must not abort the walk: a matrix missing one
		// provider is recoverable, a matrix that failed to build is not.
		if walkErr != nil {
			log.Printf("walking %s: %v", path, walkErr)

			return nil
		}

		if entry.IsDir() || entry.Name() != "schemas.json" {
			return nil
		}

		indexSchemaFile(index, path)

		return nil
	})
	if err != nil {
		log.Printf("walking providers for schemas.json: %v", err)
	}

	for _, modules := range index {
		for moduleID := range modules {
			sort.Strings(modules[moduleID])
			modules[moduleID] = dedupe(modules[moduleID])
		}
	}

	return index
}

// indexSchemaFile folds one schemas.json into the index under its provider directory.
func indexSchemaFile(index map[string]map[string][]string, path string) {
	rel, err := filepath.Rel("providers", path)
	if err != nil {
		return
	}

	providerDir := strings.ToLower(strings.Split(rel, string(filepath.Separator))[0])

	objectsByModule := readSchemaFile(path)
	if len(objectsByModule) == 0 {
		return
	}

	if _, ok := index[providerDir]; !ok {
		index[providerDir] = make(map[string][]string)
	}

	for moduleID, names := range objectsByModule {
		index[providerDir][moduleID] = append(index[providerDir][moduleID], names...)
	}
}

// readSchemaFile parses one schemas.json into object names keyed by module.
func readSchemaFile(path string) map[string][]string {
	data, err := os.ReadFile(path) //nolint:gosec // Path comes from walking the repo, not input.
	if err != nil {
		return nil
	}

	var parsed schemaFile

	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Printf("skipping %s: did not parse: %v", path, err)

		return nil
	}

	out := make(map[string][]string, len(parsed.Modules))

	for moduleID, module := range parsed.Modules {
		names := make([]string, 0, len(module.Objects))
		for objectName := range module.Objects {
			names = append(names, objectName)
		}

		// "root" is how schemas.json spells the default module; the catalog spells it "".
		if moduleID == "root" {
			moduleID = ""
		}

		out[moduleID] = append(out[moduleID], names...)
	}

	return out
}

func dedupe(values []string) []string {
	if len(values) < 2 { //nolint:mnd // Nothing to dedupe below two entries.
		return values
	}

	out := values[:1]

	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}

	return out
}

func marshalWithoutEscaping(v any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")

	err := encoder.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("encoding matrix: %w", err)
	}

	return buffer.Bytes(), nil
}
