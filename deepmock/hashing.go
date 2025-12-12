package deepmock

import (
	"encoding/binary"
	"errors"
	"hash"
	"math"

	"github.com/kaptinlin/jsonschema"
)

const (
	// int64Size is the number of bytes in an int64 (8 bytes for BigEndian encoding).
	int64Size = 8
)

// hashable represents types that can contribute their content to a hash computation.
// Types implementing this interface can be included in deterministic hash calculations
// via the hashBuilder.Hashable method.
type hashable interface {
	UpdateHash(h hash.Hash) error
}

// hashBuilder provides a fluent API for deterministic hashing of structured data.
//
// The builder ensures consistent hash values by:
//   - Using field separators (null bytes) between consecutive field values
//   - Encoding all data types in a canonical binary representation (BigEndian for numbers)
//   - Distinguishing between nil and non-nil pointer values
//   - Accumulating any errors that occur during hash operations
//
// This is particularly useful for creating stable cache keys or content identifiers
// from complex structs where field order and types need to be consistently represented.
//
// Example usage:
//
//	builder := &hashBuilder{h: sha256.New()}
//	builder.String("contact").Int64(12345).Bool(true)
//	if err := builder.Error(); err != nil {
//	    // handle error
//	}
//	hash := builder.h.Sum(nil)
type hashBuilder struct {
	h    hash.Hash // The underlying hash implementation (e.g., SHA-256, MD5)
	fc   int       // Field count - tracks number of fields added for separator logic
	errs []error   // Accumulated errors from hash write operations
}

// String adds a string value to the hash.
// The string is written as UTF-8 bytes.
func (b *hashBuilder) String(s string) *hashBuilder {
	b.incrementField()
	b.write([]byte(s))

	return b
}

// Nil adds a nil marker (byte value 0) to the hash.
// Used to explicitly represent nil values in the hash computation.
func (b *hashBuilder) Nil() *hashBuilder {
	b.incrementField()
	b.write([]byte{0})

	return b
}

// NonNil adds a non-nil marker (byte value 1) to the hash.
// Used to explicitly represent non-nil values in the hash computation.
func (b *hashBuilder) NonNil() *hashBuilder {
	b.incrementField()
	b.write([]byte{1})

	return b
}

// Rat adds a rational number (jsonschema.Rat) to the hash.
// Nil values are represented as byte 0, non-nil values as byte 1 followed by
// the string representation of the rational number.
func (b *hashBuilder) Rat(rat *jsonschema.Rat) {
	b.incrementField()

	if rat == nil {
		b.write([]byte{0})
	} else {
		b.write([]byte{1})
		b.write([]byte(rat.RatString()))
	}
}

// StringPtr adds a string pointer to the hash.
// Nil pointers are represented as byte 0, non-nil pointers as byte 1 followed by
// the string value. This ensures nil and empty string produce different hashes.
func (b *hashBuilder) StringPtr(s *string) *hashBuilder {
	b.incrementField()

	if s == nil {
		b.write([]byte{0})
	} else {
		b.write([]byte{1})
		b.write([]byte(*s))
	}

	return b
}

// Bool adds a boolean value to the hash.
// True is represented as byte 1, false as byte 0.
func (b *hashBuilder) Bool(v bool) *hashBuilder {
	b.incrementField()

	if v {
		b.write([]byte{1})
	} else {
		b.write([]byte{0})
	}

	return b
}

// BoolPtr adds a boolean pointer to the hash.
// Nil pointers are represented as {0, 0}, true as {1, 1}, and false as {1, 0}.
// This three-way distinction ensures nil, false, and true all produce different hashes.
func (b *hashBuilder) BoolPtr(v *bool) *hashBuilder {
	b.incrementField()

	switch {
	case v == nil:
		b.write([]byte{0, 0})
	case *v:
		b.write([]byte{1, 1})
	default:
		b.write([]byte{1, 0})
	}

	return b
}

// Int64 adds a 64-bit integer to the hash.
// The integer is encoded as 8 bytes in big-endian format for deterministic representation
// across different architectures.
func (b *hashBuilder) Int64(v int64) *hashBuilder {
	b.incrementField()

	bts := make([]byte, int64Size)
	binary.BigEndian.PutUint64(bts, uint64(v))

	b.write(bts)

	return b
}

// Float64 adds a 64-bit floating point number to the hash.
// The float is converted to its IEEE 754 bit representation and encoded as 8 bytes
// in big-endian format for deterministic representation.
func (b *hashBuilder) Float64(v float64) *hashBuilder {
	b.incrementField()

	bits64 := math.Float64bits(v)

	bts := make([]byte, int64Size)

	binary.BigEndian.PutUint64(bts, bits64)

	b.write(bts)

	return b
}

// Float64Ptr adds a float64 pointer to the hash.
// Nil pointers are represented as byte 0, non-nil pointers as byte 1 followed by
// the 8-byte IEEE 754 representation in big-endian format.
func (b *hashBuilder) Float64Ptr(value *float64) *hashBuilder {
	b.incrementField()

	if value == nil {
		b.write([]byte{0})

		return b
	}

	b.write([]byte{1})

	bits64 := math.Float64bits(*value)

	bts := make([]byte, int64Size)

	binary.BigEndian.PutUint64(bts, bits64)

	b.write(bts)

	return b
}

// Hashable adds a hashable type to the hash computation.
// Nil values are represented as byte 0, non-nil values as byte 1 followed by
// the result of calling UpdateHash on the hashable object.
// Any errors from UpdateHash are accumulated in the builder's error list.
func (b *hashBuilder) Hashable(h hashable) *hashBuilder {
	b.incrementField()

	if h == nil {
		b.write([]byte{0})
	} else {
		b.write([]byte{1})
	}

	if err := h.UpdateHash(b.h); err != nil {
		b.errs = append(b.errs, err)
	}

	return b
}

// Error returns any accumulated errors from hash operations.
// Returns nil if no errors occurred, a single error if only one occurred,
// or a joined error combining all errors if multiple occurred.
func (b *hashBuilder) Error() error {
	if len(b.errs) == 0 {
		return nil
	} else if len(b.errs) == 1 {
		return b.errs[0]
	}

	return errors.Join(b.errs...)
}

// incrementField adds a field separator (null byte) before each field after the first.
// This ensures that different field combinations produce different hashes.
// For example, String("ab").String("c") will hash differently than String("a").String("bc").
func (b *hashBuilder) incrementField() {
	if b.fc > 0 {
		b.h.Write([]byte{0})
	}

	b.fc++
}

// write writes bytes to the hash and accumulates any errors.
// Errors are rare with standard hash implementations but are tracked for completeness.
func (b *hashBuilder) write(bts []byte) {
	_, err := b.h.Write(bts)
	if err != nil {
		b.errs = append(b.errs, err)
	}
}
