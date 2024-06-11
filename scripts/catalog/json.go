package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/amp-labs/connectors/providers"
)

const writePerm = 0o644

func main() {
	catalog, err := providers.ReadCatalog()
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.Marshal(catalog)
	if err != nil {
		log.Fatal(err)
	}

	tempFile := "providers/catalog.json"

	err = os.WriteFile(tempFile, bytes, writePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}
