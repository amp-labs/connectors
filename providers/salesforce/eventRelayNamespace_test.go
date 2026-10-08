package salesforce

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestFullNameMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		fullName     string
		wantExpected string
		wantOK       bool
	}{
		{
			name: "namespaced full name mismatch is detected and expected name extracted",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body: []byte(`[{"message":"Full name amp_7172c2e551fc4382a034d2c27f23e8eb ` +
					`does not match the full name speedboatdev__amp_7172c2e551fc4382a034d2c27f23e8eb",` +
					`"errorCode":"FIELD_INTEGRITY_EXCEPTION"}]`),
			},
			fullName:     "amp_7172c2e551fc4382a034d2c27f23e8eb",
			wantExpected: "speedboatdev__amp_7172c2e551fc4382a034d2c27f23e8eb",
			wantOK:       true,
		},
		{
			name: "real salesforce message with trailing entity clause",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body: []byte(`[{"message":"Full name amp_d91acb347f434502a3063f436d82263a ` +
					`does not match the full name speedboatdev__amp_d91acb347f434502a3063f436d82263a ` +
					`of the entity (id: 7k2gL0000046h6PQAQ).","errorCode":"FIELD_INTEGRITY_EXCEPTION","fields":[]}]`),
			},
			fullName:     "amp_d91acb347f434502a3063f436d82263a",
			wantExpected: "speedboatdev__amp_d91acb347f434502a3063f436d82263a",
			wantOK:       true,
		},
		{
			name: "namespace containing a single underscore is captured in full",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body: []byte(`[{"message":"Full name amp_d91acb347f434502a3063f436d82263a ` +
					`does not match the full name my_np__amp_d91acb347f434502a3063f436d82263a ` +
					`of the entity (id: 7k2gL0000046h6PQAQ).","errorCode":"FIELD_INTEGRITY_EXCEPTION","fields":[]}]`),
			},
			fullName:     "amp_d91acb347f434502a3063f436d82263a",
			wantExpected: "my_np__amp_d91acb347f434502a3063f436d82263a",
			wantOK:       true,
		},
		{
			name: "trailing sentence after expected name does not bleed into match",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body: []byte(`[{"message":"Full name amp_x does not match the full name acme__amp_x. ` +
					`Please retry.","errorCode":"FIELD_INTEGRITY_EXCEPTION"}]`),
			},
			fullName:     "amp_x",
			wantExpected: "acme__amp_x",
			wantOK:       true,
		},
		{
			name: "un-namespaced supplied name alone does not match",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body:   []byte(`[{"message":"Full name amp_x is invalid","errorCode":"FIELD_INTEGRITY_EXCEPTION"}]`),
			},
			fullName:     "amp_x",
			wantExpected: "",
			wantOK:       false,
		},
		{
			name: "unrelated 400 is not a mismatch",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body:   []byte(`[{"message":"Something else went wrong","errorCode":"INVALID_FIELD"}]`),
			},
			fullName:     "amp_x",
			wantExpected: "",
			wantOK:       false,
		},
		{
			name: "non-400 status is ignored",
			err: &common.HTTPError{
				Status: http.StatusInternalServerError,
				Body:   []byte(`[{"message":"does not match the full name acme__amp_x"}]`),
			},
			fullName:     "amp_x",
			wantExpected: "",
			wantOK:       false,
		},
		{
			name: "empty full name is ignored",
			err: &common.HTTPError{
				Status: http.StatusBadRequest,
				Body:   []byte(`[{"message":"does not match the full name acme__amp_x"}]`),
			},
			fullName:     "",
			wantExpected: "",
			wantOK:       false,
		},
		{
			name:         "non-HTTP error is ignored",
			err:          errors.New("boom"),
			fullName:     "amp_x",
			wantExpected: "",
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expected, ok := fullNameMismatch(tt.err, tt.fullName)
			assert.Equal(t, ok, tt.wantOK)
			assert.Equal(t, expected, tt.wantExpected)
		})
	}
}
