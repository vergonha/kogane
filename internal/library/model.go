package library

import "unicode/utf8"

type Manga struct {
	Title                  string    `json:"Title"`
	Volumes                []string  `json:"Volumes"`
	Cover                  string    `json:"Cover"`
	Description            string    `json:"description"`
	Folder                 string    `json:"folder"`
	DescriptionLanguage    string    `json:"description_language"`
	Status                 string    `json:"status"`
	Year                   int       `json:"year"`
	ContentRating          string    `json:"content_rating"`
	OriginalLanguage       string    `json:"original_language"`
	Authors                []string  `json:"authors"`
	Artists                []string  `json:"artists"`
	Tags                   []string  `json:"tags"`
	PublicationDemographic string    `json:"publication_demographic"`
	Metadata               *metadata `json:"metadata"`
}

// metadata is the raw provider payload each entry carries alongside the
// flattened top-level fields.
type metadata struct {
	Provider string     `json:"provider"`
	Kitsu    *kitsuMeta `json:"kitsu"`
}

type kitsuMeta struct {
	Status   string   `json:"status"`
	CoverURL string   `json:"cover_url"`
	Genres   []string `json:"genres"`
	Authors  []string `json:"authors"`
	Artists  []string `json:"artists"`
}

func (m Manga) PreviewDescription() string {
	const limit = 150

	if utf8.RuneCountInString(m.Description) <= limit {
		return m.Description
	}

	runes := []rune(m.Description)
	return string(runes[:limit]) + "..."
}

func (m Manga) PreviewImageURL() string {
	if m.Metadata == nil || m.Metadata.Kitsu == nil {
		return ""
	}

	return m.Metadata.Kitsu.CoverURL
}

// fillFromMetadata copies provider metadata into the flattened fields the
// templates read, for entries where the generator left them empty.
func (m *Manga) fillFromMetadata() {
	if m.Metadata == nil || m.Metadata.Kitsu == nil {
		return
	}

	if m.Status == "" {
		m.Status = m.Metadata.Kitsu.Status
	}
	if len(m.Tags) == 0 {
		m.Tags = m.Metadata.Kitsu.Genres
	}
	if len(m.Authors) == 0 {
		m.Authors = m.Metadata.Kitsu.Authors
	}
	if len(m.Artists) == 0 {
		m.Artists = m.Metadata.Kitsu.Artists
	}
}
