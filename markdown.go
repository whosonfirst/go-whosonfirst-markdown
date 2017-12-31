package markdown

type Document struct {
	FrontMatter *FrontMatter
	Body        []byte
}

type FrontMatter struct {
	Title   string
	Excerpt string
	Image   string
	Authors []string
	Tags    []string
	Date    string
	URI     string
}
