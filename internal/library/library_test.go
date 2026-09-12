package library

import "testing"

func TestValidComponent(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"plain", "Berserk", true},
		{"volume file", "v01.pdf", true},
		{"spaces", "One Piece", true},
		{"empty", "", false},
		{"dot", ".", false},
		{"dotdot", "..", false},
		{"traversal", "../etc/passwd", false},
		{"embedded traversal", "a..b", false},
		{"forward slash", "a/b", false},
		{"backslash", `a\b`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidComponent(tt.in); got != tt.want {
				t.Errorf("ValidComponent(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
