// nolint:revive
package common

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen,gocognit, cyclop
func TestStringMap(t *testing.T) {
	t.Parallel()

	//nolint:varnamelen
	m := StringMap{
		"string": "string",
		"int":    1,
		"float":  1.1,
		"map":    map[string]any{"key": "value"},
		"bool":   true,
	}

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		want := "string"
		got, err := m.Get("string")

		require.NoError(t, err)
		assert.Equal(t, want, got)

		var want2 any = nil

		got2, err2 := m.Get("notfound")

		require.Error(t, err2)
		assert.Equal(t, want2, got2)

		want3 := map[string]any{"key": "value"}
		got3, err3 := m.Get("map")

		require.NoError(t, err3)
		assert.Equal(t, want3, got3)
	})

	t.Run("GetInt", func(t *testing.T) {
		t.Parallel()

		want := int64(1)
		got, err := m.GetInt("int")

		require.NoError(t, err)
		assert.Equal(t, want, got)

		keysWithIntValues := []string{"int"}

		for k := range m {
			got, err := m.GetInt(k)
			if !slices.Contains(keysWithIntValues, k) {
				require.Error(t, err)
				assert.Equal(t, int64(0), got)
			}
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		t.Parallel()

		want := 1.1
		got, err := m.GetFloat("float")

		require.NoError(t, err)
		assert.Equal(t, want, got) //nolint:testifylint

		keysWithFloatValues := []string{"float"}

		for k := range m {
			got, err := m.GetFloat(k)
			if !slices.Contains(keysWithFloatValues, k) {
				require.Error(t, err)
				assert.Equal(t, 0.0, got) //nolint:testifylint
			}
		}
	})

	t.Run("AsInt", func(t *testing.T) {
		t.Parallel()

		want := int64(1)
		got, err := m.AsInt("int")

		require.NoError(t, err)
		assert.Equal(t, want, got)

		want2 := int64(1)
		got2, err2 := m.AsInt("float")

		require.NoError(t, err2)
		assert.Equal(t, want2, got2)

		keysWithNumericValues := []string{"int", "float"}

		for k := range m {
			got, err := m.AsInt(k)
			if !slices.Contains(keysWithNumericValues, k) {
				require.Error(t, err)
				assert.Equal(t, int64(0), got)
			}
		}
	})

	t.Run("AsFloat", func(t *testing.T) {
		t.Parallel()

		want := 1.1
		got, err := m.AsFloat("float")

		require.NoError(t, err)
		assert.Equal(t, want, got) //nolint:testifylint

		want2 := 1.0
		got2, err2 := m.AsFloat("int")

		require.NoError(t, err2)
		assert.Equal(t, want2, got2) //nolint:testifylint

		keysWithNumericValues := []string{"int", "float"}

		for k := range m {
			got, err := m.AsFloat(k)
			if !slices.Contains(keysWithNumericValues, k) {
				require.Error(t, err)
				assert.Equal(t, 0.0, got) //nolint:testifylint
			}
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		t.Parallel()

		want := true
		got, err := m.GetBool("bool")

		require.NoError(t, err)
		assert.Equal(t, want, got)

		keysWithBooleanValues := []string{"bool"}

		for k := range m {
			got, err := m.GetBool(k)
			if !slices.Contains(keysWithBooleanValues, k) {
				require.Error(t, err)
				assert.False(t, got)
			}
		}
	})

	t.Run("GetString", func(t *testing.T) {
		t.Parallel()

		want := "string"
		got, err := m.GetString("string")

		require.NoError(t, err)
		assert.Equal(t, want, got)

		keysWithStringValues := []string{"string"}

		for k := range m {
			got, err := m.GetString(k)
			if !slices.Contains(keysWithStringValues, k) {
				require.Error(t, err)
				assert.Equal(t, "", got)
			}
		}
	})
}
