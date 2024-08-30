package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	mdata := make(map[string]any)

	def, docA, err := constructDefinitions("./scripts/openapi/marketo/metadata/asset.json", 5) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	ldef, docL, err := constructDefinitions("./scripts/openapi/marketo/metadata/mapi.json", 4) //nolint:gomnd
	if err != nil {
		panic(err)
	}

	mp, err := generateMetadata(ldef, docL, mdata)
	if err != nil {
		panic(err)
	}

	mp, err = generateMetadata(def, docA, mp)
	if err != nil {
		panic(err)
	}

	mb, err := json.Marshal(mp)
	if err != nil {
		panic(err)
	}

	if err = writefile(mb); err != nil {
		panic(err)
	}
}

func writefile(b []byte) error {
	f, err := os.Create("schemas.json")
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	fmt.Println("Successfuly written metadatas to schemas.json")

	return nil
}

func constructDefinitions(file string, length int) (map[string]string, openapi2.T, error) {
	var definitions = make(map[string]string)

	f, err := os.ReadFile(file)
	if err != nil {
		return nil, openapi2.T{}, err
	}

	var doc openapi2.T
	if err = json.Unmarshal(f, &doc); err != nil {
		return nil, openapi2.T{}, err
	}

	for k, v := range doc.Paths {
		if pathLength(k) == length && v.Get != nil {
			obj := retrieveObject(k)

			for _, j := range v.Get.Responses {
				dfn := cleanDefinitions(j.Schema.Ref)
				definitions[obj] = dfn
			}
		}
	}

	return definitions, doc, nil
}

func pathLength(path string) int {
	return len(strings.Split(path, "/"))
}

func retrieveObject(path string) string {
	s := strings.Split(path, "/")
	sWithJSON := s[len(s)-1]

	return strings.TrimSuffix(sWithJSON, ".json")
}

func cleanDefinitions(def string) string {
	s := strings.Split(def, "/")

	return s[len(s)-1]
}

func generateMetadata(objDefs map[string]string, doc openapi2.T, output map[string]any) (map[string]any, error) {
	for obj, dfn := range objDefs {
		schemas := doc.Definitions[dfn].Value.Properties

		// This is attached to the marketo
		result, err := schemas["result"].Value.JSONLookup("items")
		if err != nil {
			return nil, err
		}

		r, ok := result.(*openapi3.Ref)
		if !ok {
			return nil, fmt.Errorf("failed the assertion, the response data type is not expected") //nolint:goerr113
		}

		m := cleanDefinitions(r.Ref)
		lschems := doc.Definitions[m].Value.Properties
		var fields = make(map[string]string)

		for k := range lschems {
			fields[k] = k
		}

		output[obj] = fields
	}

	return output, nil
}

func format(m map[string]any) scrapper.ObjectMetadataResult {
	var f = &scrapper.ObjectMetadataResult{}

	var a = make(map[string]scrapper.ObjectMetadata)

	for k, v := range m {
		v, ok := v.(map[string]string)
		if !ok {
			panic(!ok)
		}

		a[k] = scrapper.ObjectMetadata{
			DisplayName: k,
			FieldsMap:   v,
		}
	}

	f.Result = a

	return *f
}
