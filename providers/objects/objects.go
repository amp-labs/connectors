package objects

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed matrix.json
var matrixJSON []byte

//nolint:gochecknoglobals
var (
	loadOnce sync.Once
	loaded   Matrix
	// errLoad holds the parse failure, if any, from the one-time load.
	errLoad error
)

// Load returns the generated matrix. It is parsed once and shared.
func Load() (Matrix, error) {
	loadOnce.Do(func() {
		errLoad = json.Unmarshal(matrixJSON, &loaded)
		if errLoad != nil {
			errLoad = fmt.Errorf("parsing embedded object matrix: %w", errLoad)
		}
	})

	return loaded, errLoad
}

// ForProvider returns a provider's entry.
func ForProvider(provider string) (ProviderObjects, bool) {
	matrix, err := Load()
	if err != nil {
		return ProviderObjects{}, false
	}

	entry, ok := matrix.Providers[provider]

	return entry, ok
}

// Answer is what the matrix can say about one provider, object and operation.
type Answer string

const (
	// AnswerSupported means the object is recorded and the operation is supported.
	AnswerSupported Answer = "supported"

	// AnswerUnsupported means the object is recorded and the operation is not supported.
	AnswerUnsupported Answer = "unsupported"

	// AnswerUnknown means the matrix cannot say. For a provider that works against
	// whatever object the credential can reach, an unlisted object may well work; for a
	// provider we hold no data on at all, nothing is known either way.
	//
	// Kept distinct from AnswerUnsupported on purpose: telling someone an object is not
	// supported when the truth is that nobody checked is the failure mode worth avoiding
	// in a matrix whose whole job is to be authoritative.
	AnswerUnknown Answer = "unknown"
)

// Supports reports whether a provider supports an operation on an object.
//
// Object names are matched case-insensitively: manifests carry the provider's casing
// ("Task"), shipped schemas are generated with whatever the provider's API returns, and a
// case mismatch reported as "unsupported" would be a wrong answer rather than a missing one.
func Supports(provider, module, object string, operation Operation) Answer {
	entry, ok := ForProvider(provider)
	if !ok {
		return AnswerUnknown
	}

	support, found := lookup(entry, module, object)
	if !found {
		if entry.Generic {
			// The provider reaches whatever the credential can. Absence says nothing.
			return AnswerUnknown
		}

		return AnswerUnsupported
	}

	if support.Supports(operation) {
		return AnswerSupported
	}

	return AnswerUnsupported
}

// lookup finds an object within a provider entry, preferring the named module and falling
// back to any module that has it.
//
// The fallback exists because callers do not always know which module an object lives in --
// a manifest names an object, not a module -- and refusing to answer over that would make
// the matrix much less useful than it needs to be.
func lookup(entry ProviderObjects, module, object string) (ObjectSupport, bool) {
	wanted := strings.ToLower(object)

	if objectsInModule, ok := entry.Objects[module]; ok {
		if support, found := findObject(objectsInModule, wanted); found {
			return support, true
		}
	}

	for moduleID, objectsInModule := range entry.Objects {
		if moduleID == module {
			continue
		}

		if support, found := findObject(objectsInModule, wanted); found {
			return support, true
		}
	}

	return ObjectSupport{}, false
}

func findObject(objectsInModule map[string]ObjectSupport, wanted string) (ObjectSupport, bool) {
	for name, support := range objectsInModule {
		if strings.ToLower(name) == wanted {
			return support, true
		}
	}

	return ObjectSupport{}, false
}
