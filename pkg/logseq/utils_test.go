package logseq_test

import (
	"testing"

	"github.com/clstb/yalms/pkg/logseq"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		content  string
		expected []string
	}{
		{"Simple [[Link]] test", []string{"Link"}},
		{"Multiple [[Link1]] and [[Link2]]", []string{"Link1", "Link2"}},
		{"Namespace [[A/B]] link", []string{"A/B"}},
		{"Empty [[]] link", nil}, // or empty slice
		{"No links here", nil},
		{"Duplicate [[Link]] and [[Link]]", []string{"Link"}},
		{"Spaces [[ Link With Spaces ]]", []string{"Link With Spaces"}},
	}

	for _, tt := range tests {
		got := logseq.ExtractLinks(tt.content)
		if len(got) != len(tt.expected) {
			t.Errorf("extractLinks(%q) count = %d, want %d", tt.content, len(got), len(tt.expected))
		}
		for i, v := range got {
			if v != tt.expected[i] {
				t.Errorf("extractLinks(%q)[%d] = %q, want %q", tt.content, i, v, tt.expected[i])
			}
		}
	}
}
