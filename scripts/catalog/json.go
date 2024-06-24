package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors/catalog"
)

const writePerm = 0o644

const countFileContents = `package generated

// This file will be updated automatically, do not edit it manually.

const ProviderCount = %d
`

func main() {
	cat, err := catalog.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.Marshal(cat)
	if err != nil {
		log.Fatal(err)
	}

	tempFile := "providers/catalog.json"

	err = os.WriteFile(tempFile, bytes, writePerm)
	if err != nil {
		log.Fatal(err)
	}

	countFile := "internal/generated/provider_count.go"
	str := fmt.Sprintf(countFileContents, len(cat))

	err = os.WriteFile(countFile, []byte(str), writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}
