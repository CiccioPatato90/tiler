package main

import "testing"

func TestIsExpanded(t *testing.T) {
	tests := []struct {
		name     string
		percents []float64
		want     bool
	}{
		{"two equal", []float64{0.5, 0.5}, false},
		{"two expanded", []float64{0.33, 0.67}, true},
		{"three old expanded", []float64{0.25, 0.5, 0.25}, true},
		{"three equal", []float64{0.33, 0.34, 0.33}, false},
		{"four old expanded", []float64{0.167, 0.5, 0.167, 0.166}, true},
		{"four equal", []float64{0.25, 0.25, 0.25, 0.25}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := Workspace{Nodes: make([]Node, len(tt.percents))}
			for i, percent := range tt.percents {
				workspace.Nodes[i].Percent = percent
			}
			if got := workspace.IsExpanded(); got != tt.want {
				t.Fatalf("IsExpanded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetWidths(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []float64
	}{
		{"two", 2, []float64{100.0 / 3, 200.0 / 3}},
		{"three", 3, []float64{25, 50, 25}},
		{"four", 4, []float64{100.0 / 6, 50, 100.0 / 6, 100.0 / 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := Workspace{Nodes: make([]Node, tt.n)}
			workspace.Nodes[1].Focused = true

			got := workspace.TargetWidths()
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("width[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEqualWidths(t *testing.T) {
	workspace := Workspace{Nodes: make([]Node, 4)}

	for i, width := range workspace.EqualWidths() {
		if width != 25 {
			t.Fatalf("width[%d] = %v, want 25", i, width)
		}
	}
}
