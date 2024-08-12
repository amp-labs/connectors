package main

import (
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
	ts := time.Now().UTC().Format(time.RFC3339)

	catalog, err := providers.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

	catalog.Timestamp = ts

	bytes, err := json.Marshal(catalog)
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
	str = fmt.Sprintf(timestampFileContents, ts)

	err = os.WriteFile(timestampFile, []byte(str), writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}
