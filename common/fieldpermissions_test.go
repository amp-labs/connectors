// nolint:revive,godoclint
package common

import "testing"

// TestFieldPermissionsIsWritable pins the helper that replaces a !ReadOnly check.
//
// The nil cases matter as much as the true ones: "the provider did not say" must not read
// as "yes". A caller deciding whether to attempt a write on an unknown field should be told
// it does not know, and act accordingly, rather than be handed a confident false positive.
func TestFieldPermissionsIsWritable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		permissions *FieldPermissions
		want        bool
	}{
		{"nil permissions are not writable", nil, false},
		{"nothing reported is not writable", &FieldPermissions{Readable: new(true)}, false},
		{
			name:        "createable only",
			permissions: &FieldPermissions{Createable: new(true), Updateable: new(false)},
			want:        true,
		},
		{
			name:        "updateable only",
			permissions: &FieldPermissions{Createable: new(false), Updateable: new(true)},
			want:        true,
		},
		{
			name:        "both",
			permissions: &FieldPermissions{Createable: new(true), Updateable: new(true)},
			want:        true,
		},
		{
			name:        "neither",
			permissions: &FieldPermissions{Createable: new(false), Updateable: new(false)},
			want:        false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.permissions.IsWritable(); got != testCase.want {
				t.Errorf("IsWritable() = %v, want %v", got, testCase.want)
			}
		})
	}
}

// TestFieldPermissionsExpressWhatReadOnlyCannot is the case for the type existing at all.
//
// Each of these three fields has ReadOnly=false -- they all look equally writable -- while
// only one of them actually accepts both a create and an update. That gap is what turns
// "can this customer write this field" into a question you can only answer by trying it.
func TestFieldPermissionsExpressWhatReadOnlyCannot(t *testing.T) {
	t.Parallel()

	setOnceAtCreation := FieldMetadata{
		ReadOnly:    new(false),
		Permissions: &FieldPermissions{Createable: new(true), Updateable: new(false)},
	}
	systemManagedUntilSet := FieldMetadata{
		ReadOnly:    new(false),
		Permissions: &FieldPermissions{Createable: new(false), Updateable: new(true)},
	}
	fullyWritable := FieldMetadata{
		ReadOnly:    new(false),
		Permissions: &FieldPermissions{Createable: new(true), Updateable: new(true)},
	}

	for _, field := range []FieldMetadata{setOnceAtCreation, systemManagedUntilSet, fullyWritable} {
		if *field.ReadOnly {
			t.Fatal("all three are indistinguishable through ReadOnly, which is the point")
		}

		if !field.Permissions.IsWritable() {
			t.Error("all three are writable in at least one direction")
		}
	}

	if !*setOnceAtCreation.Permissions.Createable || *setOnceAtCreation.Permissions.Updateable {
		t.Error("a set-once field must report createable but not updateable")
	}

	if *systemManagedUntilSet.Permissions.Createable || !*systemManagedUntilSet.Permissions.Updateable {
		t.Error("a field the provider populates must report updateable but not createable")
	}
}
