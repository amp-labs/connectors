package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors/providers"
)

const writePerm = 0o644

const countFileContents = `package generated

// This file will be updated automatically, do not edit it manually.

const ProviderCount = %d
`

func main() {
	catalog, err := providers.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

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
	str := fmt.Sprintf(countFileContents, len(catalog))

	err = os.WriteFile(countFile, []byte(str), writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}
