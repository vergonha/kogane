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

func TestHasVolume(t *testing.T) {
	m := Manga{Volumes: []string{
		"HUNTER X HUNTER VOL.01.pdf",
		"CAPÍTULOS ATUAIS/hunter-x-hunter - Cap 411.pdf",
	}}

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"flat", "HUNTER X HUNTER VOL.01.pdf", true},
		{"nested", "CAPÍTULOS ATUAIS/hunter-x-hunter - Cap 411.pdf", true},
		{"unknown", "HUNTER X HUNTER VOL.99.pdf", false},
		{"traversal", "../../etc/passwd", false},
		{"nested traversal", "CAPÍTULOS ATUAIS/../../secret.pdf", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.HasVolume(tt.in); got != tt.want {
				t.Errorf("HasVolume(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestVolumeLabel(t *testing.T) {
	if got := VolumeLabel("CAPÍTULOS ATUAIS/cap 411.pdf"); got != "cap 411" {
		t.Errorf("VolumeLabel = %q", got)
	}
}
