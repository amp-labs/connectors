package housecallpro

import (
	"testing"
)

func TestObjectReadSpecs(t *testing.T) {
	t.Parallel()

	incremental := []struct {
		object  string
		timeKey string
	}{
		{"customers", "updated_at"},
		{"estimates", "updated_at"},
		{"jobs", "updated_at"},
		{"price_book/material_categories", "updated_at"},
		{"events", "updated_at"},
	}

	for _, tt := range incremental {
		t.Run(tt.object, func(t *testing.T) {
			t.Parallel()

			spec := objectReadSpecs.Get(tt.object)
			if spec.timeKey != tt.timeKey {
				t.Fatalf("timeKey: got %q, want %q", spec.timeKey, tt.timeKey)
			}

			if !spec.supportsIncremental {
				t.Fatalf("supportsIncremental: got false, want true")
			}
		})
	}

	unlisted := []string{
		"leads",
		"lead_sources",
		"invoices",
		"employees",
		"price_book/price_forms",
		"price_book/services",
		"job_fields/job_types",
		"service_zones",
		"routes",
		"tags",
		"not_registered_in_provider_yet",
	}

	for _, object := range unlisted {
		t.Run(object, func(t *testing.T) {
			t.Parallel()

			spec := objectReadSpecs.Get(object)
			if spec.timeKey != "" || spec.supportsIncremental {
				t.Fatalf("unlisted object spec: got %+v, want zero objectReadSpec", spec)
			}
		})
	}
}
