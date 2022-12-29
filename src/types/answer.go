package types

import "html/template"

type FilteredAnswer struct {
	Upvotes    string
	IsAccepted bool

	AuthorName string
	AuthorURL  string

	Timestamp string

	Body template.HTML

	Comments []FilteredComment
}
