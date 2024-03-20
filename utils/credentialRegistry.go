// nolint: ireturn
package utils

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-playground/validator"
	"github.com/spyzhov/ajson"
)

var (
	validate                = validator.New() //nolint:gochecknoglobals
	ErrKeyNotFound          = errors.New("key not found")
	ErrWrongType            = errors.New("wrong type")
	ErrEnvVarNotSet         = errors.New("environment variable not set")
	ErrEmptyFilePathPathVar = errors.New("empty value for file path")
	ErrJSONPathNotFound     = errors.New("empty value at json path")
	ErrReaderNotFound       = errors.New("Reader not found")
	ErrCredentialNotFound   = errors.New("credential not found")
)

type CredentialsRegistry map[string]Reader

type Reader interface {
	Value() (any, error)
	Key() (string, error)
}

type EnvReaderParam struct {
	EnvName string `json:"envVar" validate:"required"`
}

type EnvReader struct {
	EnvName string `json:"params" validate:"required"`
	CredKey string `json:"string" validate:"required"`
}

func (r *EnvReader) Value() (any, error) {
	value := os.Getenv(r.EnvName)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrEnvVarNotSet, r.EnvName)
	}

	return value, nil
}

func (r *EnvReader) Key() (string, error) {
	if r.CredKey == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.EnvName)
	}

	return r.CredKey, nil
}

type ValueReader struct {
	Val     any    `json:"val"    validate:"required"`
	CredKey string `json:"string" validate:"required"`
}

func (r *ValueReader) Value() (any, error) {
	if r.Val == nil {
		return "", fmt.Errorf("%w: %s", ErrCredentialNotFound, r.CredKey)
	}

	return r.Val, nil
}

func (r *ValueReader) Key() (string, error) {
	if r.CredKey == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.Val)
	}

	return r.CredKey, nil
}

type JSONReader struct {
	FilePath string `json:"filePath" validate:"required"`
	JSONPath string `json:"jsonPath" validate:"required"`
	CredKey  string `json:"string"   validate:"required"`
}

func (r *JSONReader) Key() (string, error) {
	if r.CredKey == "" {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, r.FilePath)
	}

	return r.CredKey, nil
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

func (c CredentialsRegistry) AddReader(reader Reader) error {
	if err := validate.Struct(reader); err != nil {
		return err
	}

	key, err := reader.Key()
	if err != nil {
		return err
	}

	c[key] = reader

	return nil
}

func (c CredentialsRegistry) AddReaders(readers ...Reader) error {
	for _, reader := range readers {
		if err := validate.Struct(reader); err != nil {
			return fmt.Errorf("%w: %v", err, reader)
		}

		if err := c.AddReader(reader); err != nil {
			return err
		}
	}

	return nil
}

func (c CredentialsRegistry) Get(key string) (any, error) {
	reader, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return reader.Value()
}

func (c CredentialsRegistry) GetString(key string) (string, error) {
	reader, ok := c[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[string](reader)
}

func (c CredentialsRegistry) GetBool(key string) (bool, error) {
	reader, ok := c[key]
	if !ok {
		return false, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[bool](reader)
}

func (c CredentialsRegistry) GetInt(key string) (int, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int](reader)
}

func (c CredentialsRegistry) GetFloat64(key string) (float64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[float64](reader)
}

func (c CredentialsRegistry) GetFloat32(key string) (float32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[float32](reader)
}

func (c CredentialsRegistry) GetInt64(key string) (int64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int64](reader)
}

func (c CredentialsRegistry) GetInt32(key string) (int32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int32](reader)
}

func (c CredentialsRegistry) GetInt16(key string) (int16, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int16](reader)
}

func (c CredentialsRegistry) GetInt8(key string) (int8, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int8](reader)
}

func (c CredentialsRegistry) GetUint64(key string) (uint64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint64](reader)
}

func (c CredentialsRegistry) GetUint32(key string) (uint32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint32](reader)
}

func (c CredentialsRegistry) GetUint16(key string) (uint16, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint16](reader)
}

func (c CredentialsRegistry) GetUint8(key string) (uint8, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint8](reader)
}

func (c CredentialsRegistry) GetUint(key string) (uint, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint](reader)
}

func (c CredentialsRegistry) GetMap(key string) (map[string]*ajson.Node, error) {
	reader, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[map[string]*ajson.Node](reader)
}

func (c CredentialsRegistry) MustString(credKey string) string {
	str, err := c.GetString(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return str
}

func (c CredentialsRegistry) MustBool(credKey string) bool {
	b, err := c.GetBool(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (c CredentialsRegistry) MustInt(credKey string) int {
	i, err := c.GetInt(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustFloat64(credKey string) float64 {
	f, err := c.GetFloat64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func (c CredentialsRegistry) MustFloat32(credKey string) float32 {
	f, err := c.GetFloat32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func (c CredentialsRegistry) MustInt64(credKey string) int64 {
	i, err := c.GetInt64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustInt32(credKey string) int32 {
	i, err := c.GetInt32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustInt16(credKey string) int16 {
	i, err := c.GetInt16(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustInt8(credKey string) int8 {
	i, err := c.GetInt8(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustUint64(credKey string) uint64 {
	i, err := c.GetUint64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustUint32(credKey string) uint32 {
	i, err := c.GetUint32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustUint16(credKey string) uint16 {
	i, err := c.GetUint16(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustUint8(credKey string) uint8 {
	i, err := c.GetUint8(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c CredentialsRegistry) MustUint(credKey string) uint {
	i, err := c.GetUint(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func NewCredentialsRegistry() CredentialsRegistry {
	return make(CredentialsRegistry)
}

func getFromReader[T any](reader Reader) (T, error) {
	var of T

	val, err := reader.Value()
	if err != nil {
		return of, err
	}

	return get[T](val)
}

func get[T any](val any) (T, error) {
	var of T

	v, ok := (val).(T)
	if !ok {
		return of, fmt.Errorf("%w. expected %T, got %T", ErrWrongType, of, val)
	}

	return v, nil
}
