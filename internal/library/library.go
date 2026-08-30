package library

import (
	"encoding/json"
	"os"
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

	return &Service{mangas: mangas}, nil
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
