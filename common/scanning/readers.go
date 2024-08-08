package scanning

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spyzhov/ajson"
)

// Reader gets data from source specific to implementation as a key-value pair.
type Reader interface {
	Value() (any, error)
	Key() (string, error)
}

type EnvReader struct {
	KeyName string `json:"string" validate:"required"`
	EnvName string `json:"params" validate:"required"`
}

func (r *EnvReader) Value() (any, error) {
	value := os.Getenv(r.EnvName)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrEnvVarNotSet, r.EnvName)
	}

	return value, nil
}

func (r *EnvReader) Key() (string, error) {
	if r.KeyName == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.EnvName)
	}

	return r.KeyName, nil
}

type ValueReader struct {
	KeyName string `json:"string" validate:"required"`
	Val     any    `json:"val"    validate:"required"`
}

func (r *ValueReader) Value() (any, error) {
	if r.Val == nil {
		return "", fmt.Errorf("%w: %s", ErrValueNotFound, r.KeyName)
	}

	return r.Val, nil
}

func (r *ValueReader) Key() (string, error) {
	if r.KeyName == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.Val)
	}

	return r.KeyName, nil
}

type JSONReader struct {
	KeyName  string `json:"string"   validate:"required"`
	FilePath string `json:"filePath" validate:"required"`
	JSONPath string `json:"jsonPath" validate:"required"`
}

func (r *JSONReader) Key() (string, error) {
	if r.KeyName == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.FilePath)
	}

	return r.KeyName, nil
}

func (r *JSONReader) Value() (any, error) {
	data, err := os.ReadFile(r.FilePath)
	if err != nil {
		slog.Error("Error opening creds.json", "error", err)

		return nil, err
	}

	credsMap, err := ajson.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	list, err := credsMap.JSONPath(r.JSONPath)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 || list[0] == nil {
		return nil, fmt.Errorf("%w: %s", ErrJSONPathNotFound, r.JSONPath)
	}

	return list[0].Value()
}
