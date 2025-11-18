package graphql

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

// PaginationParameter holds a flexible set of fields to support both offset-based and cursor-based
// pagination in GraphQL queries.
//
// This struct can be passed into a GraphQL template to dynamically populate pagination parameters.
//
// If any expected pagination parameter is not present in this struct, you can extend it
// by adding the required fields as needed based on the GraphQL schema or API design.
type PaginationParameter struct {
	Limit, Skip, Offset, First, Last, Page, PageSize        int
	HasNextPage, HasPreviousPage                            bool
	StartCursor, EndCursor, Before, After, FromDate, ToDate string
}

// Operation loads and renders a GraphQL query or mutation from a template file.
//
// It reads a .graphql template from the embedded filesystem and executes it using the provided `data`.
// The `data` parameter can include pagination values (e.g., limit, offset, cursors) for queries,
// as well as input payloads for mutations, enabling dynamic and reusable GraphQL operations.
func Operation(queryFS embed.FS, queryType, queryName string, data any) (string, error) {
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
