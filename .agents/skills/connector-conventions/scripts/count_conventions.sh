#!/bin/sh
# count_conventions.sh [connectors repo] [ref]: how many connectors use each pattern, read from git
# (no checkout needed). Counts are context for a judgement, never the judgement: a common pattern can
# be wrong, and a new one can be the maintainers' direction.
REPO=${1:-.}; R=${2:-origin/main}
cd "$REPO" || exit 1
in_providers() {  # connectors (provider dirs) whose non-test Go matches the regex
  git grep -l -E "$1" "$R" -- 'providers/*/*.go' ':!providers/*/*_test.go' \
    | sed -E "s#^$R:providers/([^/]+)/.*#\1#" | sort -u
}
n() { in_providers "$1" | wc -l | tr -d ' '; }
names() { in_providers "$1" | tr '\n' ' '; }
echo "ref: $(git log -1 --format='%h %cs' "$R")"
echo "connectors (providers/*/connector.go): $(git ls-tree -r --name-only "$R" providers | grep -cE '^providers/[^/]+/connector.go$')"
echo
echo "## construction and metadata"
echo "components.Init( (current):              $(n 'components\.Init\(')"
echo "components.Initialize( (deprecated):     $(n 'components\.Initialize\(')"
echo "schema provider (components.SchemaProvider): $(n 'components\.SchemaProvider')"
echo "  NewOpenAPISchemaProvider:  $(n 'schema\.NewOpenAPISchemaProvider')"
echo "  NewObjectSchemaProvider:   $(n 'schema\.NewObjectSchemaProvider')"
echo "  NewCompositeSchemaProvider: $(n 'schema\.NewCompositeSchemaProvider')"
echo "empty endpoint registry (allows every object): $(n 'NewEmptyEndpointRegistry')"
echo "hand-written ListObjectMetadata:        $(n 'func \(c \*Connector\) ListObjectMetadata')"
echo "  of which serve internal/objectcatalog: $(names 'Catalog\.Select')"
echo
echo "## time filters"
echo "Since/Until filtered in code (readhelper.MakeTimeFilterFunc): $(n 'MakeTimeFilterFunc')"
echo "  Unordered:          $(names 'readhelper\.Unordered')"
echo "  ChronologicalOrder: $(names 'readhelper\.ChronologicalOrder')"
echo "  ReverseOrder:       $(names 'readhelper\.ReverseOrder')"
echo "inclusive boundary (readhelper.NewTimeBoundary): $(n 'NewTimeBoundary')"
echo "exclusive Since (ExcludeSince): $(names 'ExcludeSince')"
echo
echo "## server's concerns kept out of connectors"
echo "pacing or rate limiters in a connector: $(names 'rate\.NewLimiter|ratelimit\.|time\.Tick\(|[Pp]acer')"
echo "sleeps, backoff or retry in a connector: $(names 'time\.Sleep|[Bb]ackoff|retry\.')"
echo
echo "## tests and files"
echo "tests using mockserver: $(git grep -l 'mockserver\.' "$R" -- 'providers/*/*_test.go' | sed -E "s#^$R:providers/([^/]+)/.*#\1#" | sort -u | wc -l | tr -d ' ')"
echo "file names in providers/<name>/ (top 15):"
git ls-tree -r --name-only "$R" providers | grep -E '^providers/[^/]+/[^/]+\.go$' | grep -v _test \
  | sed -E 's#providers/[^/]+/##' | sort | uniq -c | sort -rn | head -15
