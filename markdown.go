package markdown

type Document struct {
	FrontMatter *FrontMatter
	Body        []byte
}

type FrontMatter struct {
	Title     string
	Excerpt   string
	Image     string
	Layout    string
	Category  string
	Authors   []string
	Tags      []string
	Date      string
	URI       string
	Published bool
}

func EmptyFrontMatter() *FrontMatter {

	fm := FrontMatter{
		Title:     "",
		Excerpt:   "",
		Image:     "",
		Layout:    "",
		Category:  "",
		Published: false,
		Authors:   make([]string, 0),
		Tags:      make([]string, 0),
		Date:      "",
		URI:       "",
	}

	return &fm
}
