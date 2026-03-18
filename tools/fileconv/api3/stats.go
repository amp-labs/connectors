// nolint:forbidigo,godoclint
package api3

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

// nolint:gochecknoglobals
var (
	// url path to list of objects
	statsObjectsWithMultipleArrays = datautils.Map[string, []string]{}
	// set of objects
	statsObjectsWithNoArrays = datautils.Set[string]{}
	// array field name to list of objects
	statsObjectsWithAutoSelectedArrays = datautils.UniqueLists[string, string]{}
)

func PrintObjectsWithMultipleArrays() {
	keys := statsObjectsWithMultipleArrays.Keys()

	if len(keys) == 0 {
		return
	}

	fmt.Println("====================================================================")
	fmt.Println("Schemas which hold MULTIPLE array properties")
	fmt.Println("====================================================================")

	sort.Strings(keys)

	for _, key := range keys {
		arrays := statsObjectsWithMultipleArrays[key]
		sort.Strings(arrays)
		fmt.Printf("\"%v\"\n\t%v\n", key, strings.Join(arrays, "\n\t"))
	}
}

func PrintObjectsWithNoArrays() {
	objects := statsObjectsWithNoArrays.List()
	if len(objects) == 0 {
		return
	}

	fmt.Println("====================================================================")
	fmt.Println("Schemas that have NO array properties")
	fmt.Println("====================================================================")

	sort.Strings(objects)

	for _, objectURL := range objects {
		fmt.Printf("\"%v\"\n", objectURL)
	}
}

func PrintObjectsWithAutoSelectedArrays() {
	arrayFields := statsObjectsWithAutoSelectedArrays.GetBuckets()

	if len(arrayFields) == 0 {
		return
	}

	fmt.Println("====================================================================")
	fmt.Println("Keys holding schema for the following endpoints, grouped by response key")
	fmt.Println("====================================================================")

	sort.Strings(arrayFields)

	for _, key := range arrayFields {
		objects := statsObjectsWithAutoSelectedArrays[key].List()
		sort.Strings(objects)
		fmt.Printf("key: [%v], applicable for paths:\n\t%v\n", key, strings.Join(objects, "\n\t"))
	}
}
