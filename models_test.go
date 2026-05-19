package main

import "testing"

func TestSearchModels(t *testing.T) {
	tests := []struct {
		query string
		want  int
	}{
		{"", len(AllModels())},
		{"opus", 2},   // claude-opus-4-7 + openrouter variant
		{"sonnet", 2}, // claude-sonnet-4-6 + openrouter variant
		{"haiku", 2},  // claude-haiku + openrouter variant
		{"gemini", 1},
		{"nonexistent", 0},
	}
	for _, tt := range tests {
		got := SearchModels(tt.query)
		if len(got) != tt.want {
			t.Errorf("SearchModels(%q) count = %d, want %d", tt.query, len(got), tt.want)
		}
	}
}

func TestAllModels(t *testing.T) {
	models := AllModels()
	if len(models) == 0 {
		t.Error("AllModels returned empty")
	}
	// Check for duplicates
	seen := make(map[string]bool)
	for _, m := range models {
		if seen[m.ID] {
			t.Errorf("duplicate model ID: %s", m.ID)
		}
		seen[m.ID] = true
	}
}
