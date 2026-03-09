#!/bin/bash
INPUT="default.yaml"
OUTPUT="${INPUT%.yaml}_clean.yaml"

(
yq 'del(.tags)' "$INPUT" | \
yq 'del(.paths[].get    | (.summary, .operationId, .parameters, .description, .tags, .responses.default, .externalDocs.description, ."x-ms-docs-operation-type", ."x-ms-docs-grouped-path"))' - | \
yq 'del(.paths[].post   | (.summary, .operationId, .parameters, .description, .tags, .responses.default, .externalDocs.description, ."x-ms-docs-operation-type", ."x-ms-docs-grouped-path", .requestBody))' - | \
yq 'del(.paths[].patch  | (.summary, .operationId, .parameters, .description, .tags, .responses.default, .externalDocs.description, ."x-ms-docs-operation-type", ."x-ms-docs-grouped-path", .requestBody))' - | \
yq 'del(.paths[].put    | (.summary, .operationId, .parameters, .description, .tags, .responses.default, .externalDocs.description, ."x-ms-docs-operation-type", ."x-ms-docs-grouped-path", .requestBody))' - | \
yq 'del(.paths[].delete | (.summary, .operationId, .parameters, .description, .tags, .responses.default, .externalDocs.description, ."x-ms-docs-operation-type", ."x-ms-docs-grouped-path"))' - | \
yq 'del(.components.parameters)' - | \
yq 'del(.components.examples)' - | \
yq 'del(.components.requestBodies)' - | \
yq 'del(.components.schemas[].allOf[].description)' - | \
yq 'del(.components.schemas[].description)' - | \
yq 'del(.components.schemas[].allOf[].properties[] | (.description, .nullable, .pattern, .maximum, .minimum, .format))' - | \
yq 'del(.components.schemas[].properties[] | (.description, .nullable, .pattern, .maximum, .minimum, .format))' - | \
yq 'del(.paths[].description, .paths[]."x-ms-docs-grouped-path")' -
) > "$OUTPUT"

echo "Cleaned: $OUTPUT"

yq '.paths | keys | .[] | select(test("\{|\}") | not) | .[] style="double"' "$OUTPUT" > url_paths.yaml
echo "URL endpoints without identifiers: url_paths.yaml"

awk '/components:/{flag=1} flag&&/^[^ ]/&&!/components:/{flag=0} flag' default_clean.yaml > components_only.yaml
echo "Components only: components_only.yaml"
