package graphql

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

type PaginationParameter struct {
	Limit, Skip, Offset, First, Last, Page, PageSize int
	Before, After, HasNextPage, HasPreviousPage      bool
	StartCursor, EndCursor                           string
}

// GraphQLOperation loads and renders a GraphQL operation (query or mutation) from a template file.
// This function supports dynamic GraphQL operations by loading a .graphql template from the embedded
// filesystem and executing it with provided data. It supports injecting pagination parameters for queries
// (such as limit, offset, or cursor-based pagination) as well as input payloads for mutations
// through data parameter.
func GraphQLOperation(queryFS embed.FS, queryType, queryName string, data any) (string, error) {
	filePath := fmt.Sprintf("graphql/%s_%s.graphql", queryType, queryName)

	queryBytes, err := queryFS.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(queryName).Parse(string(queryBytes))
	if err != nil {
		return "", err
	}

	var queryBuf bytes.Buffer

	err = tmpl.Execute(&queryBuf, data)
	if err != nil {
		return "", err
	}

	return queryBuf.String(), nil
}
