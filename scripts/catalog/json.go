package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/amp-labs/connectors/providers"
)

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

	err = os.WriteFile(tempFile, bytes, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Catalog successfully written to: %s\n", tempFile)
}
