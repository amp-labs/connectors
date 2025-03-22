package openapi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type GraphqlIntrospection struct {
	Data struct {
		Schema struct {
			Types []struct {
				Name   string `json:"name"`
				Kind   string `json:"kind"`
				Fields []struct {
					Name string `json:"name"`
					Type struct {
						Name   *string `json:"name"`
						Kind   string  `json:"kind"`
						OfType *struct {
							Name *string `json:"name"`
						} `json:"ofType"`
					} `json:"type"`
				} `json:"fields"`
			} `json:"types"`
		} `json:"__schema"`
	} `json:"data"`
}

func graphqlToOpenAPI(introspectionJSON []byte) (map[string]interface{}, error) {
	var introspection GraphqlIntrospection
	err := json.Unmarshal(introspectionJSON, &introspection)
	if err != nil {
		return nil, err
	}

	openapiSchema := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "GraphQL API",
			"version": "1.0.0",
		},
		"paths":      map[string]interface{}{},
		"components": map[string]interface{}{"schemas": map[string]interface{}{}},
	}

	for _, typeDef := range introspection.Data.Schema.Types {
		if typeDef.Kind == "OBJECT" && typeDef.Name != "" {
			objectName := typeDef.Name
			if objectName == "Query" || objectName == "Mutation" {
				for _, field := range typeDef.Fields {
					operationID := field.Name
					httpMethod := "get"
					if objectName == "Mutation" {
						httpMethod = "post"
					}

					path := fmt.Sprintf("/%s", operationID)
					openapiSchema["paths"].(map[string]interface{})[path] = map[string]interface{}{
						httpMethod: map[string]interface{}{
							"operationId": operationID,
							"responses": map[string]interface{}{
								"200": map[string]interface{}{
									"description": "Successful response",
									"content": map[string]interface{}{
										"application/json": map[string]interface{}{
											"schema": map[string]interface{}{
												"$ref": func() string {
													if field.Type.Name != nil {
														return fmt.Sprintf("#/components/schemas/%s", *field.Type.Name)
													}
													return ""
												}(),
											},
										},
									},
								},
							},
						},
					}
				}
			} else {
				schema := map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
				for _, field := range typeDef.Fields {
					fieldName := field.Name
					fieldType := field.Type
					schema["properties"].(map[string]interface{})[fieldName] = map[string]interface{}{
						"type": mapGraphqlTypeToOpenAPI(fieldType),
					}
				}
				openapiSchema["components"].(map[string]interface{})["schemas"].(map[string]interface{})[objectName] = schema
			}
		}
	}

	return openapiSchema, nil
}

func mapGraphqlTypeToOpenAPI(graphqlType struct {
	Name   *string `json:"name"`
	Kind   string  `json:"kind"`
	OfType *struct {
		Name *string `json:"name"`
	} `json:"ofType"`
}) string {
	if graphqlType.Name != nil {
		switch *graphqlType.Name {
		case "String":
			return "string"
		case "Int":
			return "integer"
		case "Boolean":
			return "boolean"
		case "ID":
			return "string"
		}
	}

	if graphqlType.Kind == "LIST" {
		return "array"
	}

	if graphqlType.Name != nil {
		return *graphqlType.Name
	}

	return "object"
}

func main() {
	// Read introspection JSON from file
	introspectionJSON, err := os.ReadFile("./introspectionQuery.graphql")
	fmt.Println(string(introspectionJSON))
	if err != nil {
		log.Fatalf("Error reading introspection file: %v", err)
	}

	openapiSchema, err := graphqlToOpenAPI(introspectionJSON)
	if err != nil {
		log.Fatalf("Error generating OpenAPI schema: %v", err)
	}

	openapiJSON, err := json.MarshalIndent(openapiSchema, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling OpenAPI schema: %v", err)
	}

	fmt.Println(string(openapiJSON))
	err = os.WriteFile("openapi.json", openapiJSON, 0644)
	if err != nil {
		log.Fatalf("Error writing OpenAPI schema to file: %v", err)
	}
}
