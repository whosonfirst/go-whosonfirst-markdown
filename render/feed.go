package render

import (
       "text/template"
)

type FeedOptions struct {
	Format   string
	Items	 int
	Templates *template.Template
}

func DefaultFeedOptions() *FeedOptions {

	opts := FeedOptions{
		Format:   "rss",
		Items:	  10,
		Templates: nil,
	}

	return &opts
}
