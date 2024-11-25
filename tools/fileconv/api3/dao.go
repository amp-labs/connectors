package api3

import (
	"log/slog"
	"strings"
)

type urlPath struct {
	path string
}

func newURLPath(path string) *urlPath {
	return &urlPath{path: path}
}

func (p urlPath) objectName(registry map[string]string) string {
	objectName, ok := registry[p.path]
	if ok {
		return objectName
	}

	// Registry didn't include this URL path.
	// We need to do some processing to infer ObjectName from the URL path.
	if p.hasIdentifiers() {
		slog.Warn("cannot infer object name for path with identifiers", "path", p.path)

		return p.path
	}

	// The last URL part is the ObjectName describing this REST resource.
	parts := strings.Split(p.path, "/")

	return parts[len(parts)-1]
}

func (p urlPath) hasIdentifiers() bool {
	return strings.Contains(p.path, "{")
}

func (p urlPath) String() string {
	return p.path
}
