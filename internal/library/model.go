package library

type Manga struct {
	Title               string   `json:"Title"`
	Volumes             []string `json:"Volumes"`
	Cover               string   `json:"Cover"`
	Description         string   `json:"description"`
	MangaDexID          string   `json:"mangadex_id"`
	Folder              string   `json:"folder"`
	DescriptionLanguage string   `json:"description_language"`
	Status              string   `json:"status"`
	Year                int      `json:"year"`
	ContentRating       string   `json:"content_rating"`
	OriginalLanguage    string   `json:"original_language"`
	LastVolume          string   `json:"last_volume"`
	LastChapter         string   `json:"last_chapter"`
}
