package library

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
)

type Service struct {
	mangas []Manga
}

func Load(path string) (*Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mangas []Manga
	if err := json.Unmarshal(data, &mangas); err != nil {
		return nil, err
	}

	slices.SortStableFunc(mangas, func(a, b Manga) int {
		if a.Cover != "" && b.Cover == "" {
			return -1
		}
		if a.Cover == "" && b.Cover != "" {
			return 1
		}
		return 0
	})

	return &Service{
		mangas: mangas,
	}, nil
}

func (s *Service) All() []Manga {
	return s.mangas
}

func ValidComponent(value string) bool {
	if value == "" {
		return false
	}

	if value == "." || value == ".." {
		return false
	}

	return !strings.Contains(value, "..") &&
		!strings.ContainsAny(value, `/\`)
}
