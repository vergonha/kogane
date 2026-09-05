package library

type Manga struct {
	Title                  string        `json:"Title"`
	Volumes                []string      `json:"Volumes"`
	Cover                  string        `json:"Cover"`
	Description            string        `json:"description"`
	MangaDexID             string        `json:"mangadex_id"`
	Folder                 string        `json:"folder"`
	DescriptionLanguage    string        `json:"description_language"`
	Status                 string        `json:"status"`
	Year                   int           `json:"year"`
	ContentRating          string        `json:"content_rating"`
	OriginalLanguage       string        `json:"original_language"`
	LastVolume             string        `json:"last_volume"`
	LastChapter            string        `json:"last_chapter"`
	Completed              bool          `json:"completed"`
	Authors                []string      `json:"authors"`
	Artists                []string      `json:"artists"`
	Tags                   []string      `json:"tags"`
	PublicationDemographic string        `json:"publication_demographic"`
	MangaDex               *mangaDexMeta `json:"mangadex"`
}

// mangaDexMeta mirrors the raw MangaDex API shape some library.json entries
// still carry their metadata in, instead of the flattened top-level fields.
type mangaDexMeta struct {
	Tags                   []mangaDexNamedItem `json:"tags"`
	Authors                []mangaDexNamedItem `json:"authors"`
	Artists                []mangaDexNamedItem `json:"artists"`
	PublicationDemographic string              `json:"publication_demographic"`
}

type mangaDexNamedItem struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

func names(items []mangaDexNamedItem) []string {
	var out []string
	for _, item := range items {
		if item.Name != nil && *item.Name != "" {
			out = append(out, *item.Name)
		}
	}
	return out
}

// fillFromMangaDex populates the flattened metadata fields from the nested
// "mangadex" object when the top-level fields weren't already set.
func (m *Manga) fillFromMangaDex() {
	if m.MangaDex == nil {
		return
	}

	if len(m.Tags) == 0 {
		m.Tags = names(m.MangaDex.Tags)
	}
	if len(m.Authors) == 0 {
		m.Authors = names(m.MangaDex.Authors)
	}
	if len(m.Artists) == 0 {
		m.Artists = names(m.MangaDex.Artists)
	}
	if m.PublicationDemographic == "" {
		m.PublicationDemographic = m.MangaDex.PublicationDemographic
	}
}
