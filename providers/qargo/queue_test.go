package qargo

import (
	"testing"

	"github.com/go-test/deep"
)

func TestQueuePage(t *testing.T) {
	t.Parallel()

	ids := []string{"a", "b", "c", "d", "e"}

	tests := []struct {
		name     string
		ids      []string
		lastRead string
		size     int
		page     []string
		next     string
	}{
		{name: "First page", ids: ids, size: 2, page: []string{"a", "b"}, next: "b"},
		{name: "Middle page", ids: ids, lastRead: "b", size: 2, page: []string{"c", "d"}, next: "d"},
		{name: "Last page is short", ids: ids, lastRead: "d", size: 2, page: []string{"e"}},
		{name: "Last page is full", ids: ids[:4], lastRead: "b", size: 2, page: []string{"c", "d"}},
		{name: "Task read last was completed", ids: []string{"a", "c", "d"}, lastRead: "b", size: 2,
			page: []string{"c", "d"}},
		{name: "Nothing after the last read", ids: ids, lastRead: "e", size: 2},
		{name: "Empty queue", size: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			page, next := queuePage(tt.ids, tt.lastRead, tt.size)

			if diff := deep.Equal(page, tt.page); len(page)+len(tt.page) != 0 && diff != nil {
				t.Errorf("page: %v", diff)
			}

			if next != tt.next {
				t.Errorf("next: got %q, want %q", next, tt.next)
			}
		})
	}
}
