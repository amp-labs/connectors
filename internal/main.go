package main

import (
	"fmt"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep"
)

type parameters struct {
	paramsbuilder.Client
	*paramsbuilder.Workspace
}

func main() {
	var x parameters
	a, _ := deep.ExtractCatalogVariables(x)

	fmt.Println(a)

	b, _ := deep.ExtractHTTPClient(x)

	fmt.Println(b)
}
