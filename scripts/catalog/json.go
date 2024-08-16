package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amp-labs/connectors/providers"
)

const writePerm = 0o644

const countFileContents = `package generated

// This file will be updated automatically, do not edit it manually.

const ProviderCount = %d
`

const timestampFileContents = `package generated

// This file will be updated automatically, do not edit it manually.

const Timestamp = %q
`

func main() {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	catalog, err := providers.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

	catalog.Timestamp = timestamp

	bytes, err := MarshalWithoutEscaping(catalog)
	if err != nil {
		log.Fatal(err)
	}

	tempFile := "internal/generated/catalog.json"

	err = os.WriteFile(tempFile, bytes, writePerm)
	if err != nil {
		log.Fatal(err)
	}

	countFile := "internal/generated/provider_count.go"
	str := fmt.Sprintf(countFileContents, len(catalog.Catalog))

	log.Printf("Provider count: %d\n", len(catalog.Catalog))

	err = os.WriteFile(countFile, []byte(str), writePerm)
	if err != nil {
		log.Fatal(err)
	}

	timestampFile := "internal/generated/timestamp.go"
	str = fmt.Sprintf(timestampFileContents, timestamp)

	err = os.WriteFile(timestampFile, []byte(str), writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}

// MarshalWithoutEscaping marshals an object into JSON without escaping HTML characters (e.g. <, >, &)
func MarshalWithoutEscaping(v any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	// The Encode method adds a newline at the end, so we need to trim it
	// Source: https://go.dev/src/encoding/json/stream.go (Line 215)
	return bytes.TrimSpace(buffer.Bytes()), nil
}
