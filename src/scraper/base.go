package scraper

import "anonymousoverflow/src/types"

type ViewQuestionInputs struct {
	QuestionID    string
	QuestionTitle string
	SortValue     string
	Sub           string
}

type QuestionsScraper interface {
	GetQuestion(ViewQuestionInputs) (types.FilteredQuestion, []types.FilteredAnswer, error)
}
