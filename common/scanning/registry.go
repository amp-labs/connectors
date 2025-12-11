package scanning

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator"
	"github.com/spyzhov/ajson"
)

var (
	validate = validator.New() //nolint:gochecknoglobals

	ErrKeyNotFound      = errors.New("key not found")
	ErrValueNotFound    = errors.New("value not found")
	ErrWrongType        = errors.New("wrong type")
	ErrEnvVarNotSet     = errors.New("environment variable not set")
	ErrJSONPathNotFound = errors.New("empty value at json path")
	ErrReaderNotFound   = errors.New("reader not found")
)

// Registry of readers capable of scanning data of different type from file, env sources.
type Registry map[string]Reader

func NewRegistry() Registry {
	return make(Registry)
}

func (c Registry) AddReader(reader Reader) error {
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

func (c Registry) AddReaders(readers ...Reader) error {
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

func (c Registry) Get(key string) (any, error) {
	reader, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return reader.Value()
}

func (c Registry) GetString(key string) (string, error) {
	reader, ok := c[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[string](reader)
}

func (c Registry) GetBool(key string) (bool, error) {
	reader, ok := c[key]
	if !ok {
		return false, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[bool](reader)
}

func (c Registry) GetInt(key string) (int, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int](reader)
}

func (c Registry) GetFloat64(key string) (float64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[float64](reader)
}

func (c Registry) GetFloat32(key string) (float32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[float32](reader)
}

func (c Registry) GetInt64(key string) (int64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int64](reader)
}

func (c Registry) GetInt32(key string) (int32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int32](reader)
}

func (c Registry) GetInt16(key string) (int16, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int16](reader)
}

func (c Registry) GetInt8(key string) (int8, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[int8](reader)
}

func (c Registry) GetUint64(key string) (uint64, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint64](reader)
}

func (c Registry) GetUint32(key string) (uint32, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint32](reader)
}

func (c Registry) GetUint16(key string) (uint16, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint16](reader)
}

func (c Registry) GetUint8(key string) (uint8, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint8](reader)
}

func (c Registry) GetUint(key string) (uint, error) {
	reader, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[uint](reader)
}

func (c Registry) GetMap(key string) (map[string]*ajson.Node, error) {
	reader, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrReaderNotFound, key)
	}

	return getFromReader[map[string]*ajson.Node](reader)
}

func (c Registry) MustString(credKey string) string {
	str, err := c.GetString(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return str
}

func (c Registry) MustBool(credKey string) bool {
	b, err := c.GetBool(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (c Registry) MustInt(credKey string) int {
	i, err := c.GetInt(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustFloat64(credKey string) float64 {
	f, err := c.GetFloat64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func (c Registry) MustFloat32(credKey string) float32 {
	f, err := c.GetFloat32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func (c Registry) MustInt64(credKey string) int64 {
	i, err := c.GetInt64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustInt32(credKey string) int32 {
	i, err := c.GetInt32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustInt16(credKey string) int16 {
	i, err := c.GetInt16(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustInt8(credKey string) int8 {
	i, err := c.GetInt8(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustUint64(credKey string) uint64 {
	i, err := c.GetUint64(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustUint32(credKey string) uint32 {
	i, err := c.GetUint32(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustUint16(credKey string) uint16 {
	i, err := c.GetUint16(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustUint8(credKey string) uint8 {
	i, err := c.GetUint8(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func (c Registry) MustUint(credKey string) uint {
	i, err := c.GetUint(credKey)
	if err != nil {
		log.Fatal(err)
	}

	return i
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
