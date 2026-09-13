package library

import (
	"encoding/json"
	"os"
	"path"
	"slices"
	"strings"
)

type Library []Manga

func Load(path string) (Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mangas Library
	if err := json.Unmarshal(data, &mangas); err != nil {
		return nil, err
	}

	for i := range mangas {
		mangas[i].fillFromMangaDex()
	}

	// Entries with a cover render first on the dashboard grid.
	slices.SortStableFunc(mangas, func(a, b Manga) int {
		if a.Cover != "" && b.Cover == "" {
			return -1
		}
		if a.Cover == "" && b.Cover != "" {
			return 1
		}
		return 0
	})

	return mangas, nil
}

func (l Library) ByTitle(title string) (Manga, bool) {
	for _, m := range l {
		if m.Title == title {
			return m, true
		}
	}

	return Manga{}, false
}

// ValidComponent reports whether value is safe to use as a single path segment
// of an object key: no traversal, no separators.
func ValidComponent(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}

	return !strings.Contains(value, "..") &&
		!strings.ContainsAny(value, `/\`)
}

// HasVolume reports whether vol is one of the manga's volume keys. Volume
// paths may contain nested folders, so the only safe check is membership in
// the library entry itself, not path-shape validation.
func (m Manga) HasVolume(vol string) bool {
	return slices.Contains(m.Volumes, vol)
}

// VolumeLabel is the display name of a volume key: volumes can live in nested
// folders, but only the file name is worth showing.
func VolumeLabel(vol string) string {
	return strings.TrimSuffix(path.Base(vol), ".pdf")
}
