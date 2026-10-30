package core

import "testing"

func TestBuildSOQLEqCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldName string
		value     any
		want      string
	}{
		{
			name:      "String value is quoted",
			fieldName: "Email",
			value:     "jane@example.com",
			want:      "Email = 'jane@example.com'",
		},
		{
			name:      "Single quotes cannot close the string literal",
			fieldName: "email",
			value:     "x' OR email != 'x",
			want:      `email = 'x\' OR email != \'x'`,
		},
		{
			name:      "Backslash cannot escape the closing quote",
			fieldName: "email",
			value:     `x\' OR email != \'x`,
			want:      `email = 'x\\\' OR email != \\\'x'`,
		},
		{
			name:      "Boolean is unquoted",
			fieldName: "IsDeleted",
			value:     false,
			want:      "IsDeleted = false",
		},
		{
			name:      "Number is unquoted",
			fieldName: "NumberOfEmployees",
			value:     float64(42),
			want:      "NumberOfEmployees = 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := BuildSOQLEqCondition(tt.fieldName, tt.value); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
