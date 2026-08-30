package library

type Manga struct {
	Title       string   `json:"title"`
	Volumes     []string `json:"volumes"`
	Cover       string   `json:"cover"`
	Description string   `json:"description"`
}
