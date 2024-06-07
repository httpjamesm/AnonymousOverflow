package routes

import (
	"anonymousoverflow/config"
	"anonymousoverflow/src/utils"
	"fmt"
	"html"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"

	"anonymousoverflow/src/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var codeBlockRegex = regexp.MustCompile(`(?s)<pre><code>(.+?)<\/code><\/pre>`)
var questionCodeBlockRegex = regexp.MustCompile(`(?s)<pre class=".+"><code( class=".+")?>(.+?)</code></pre>`)

var soSortValues = map[string]string{
	"votes":    "scoredesc",
	"trending": "trending",
	"newest":   "modifieddesc",
	"oldest":   "createdasc",
}

func ViewQuestion(c *gin.Context) {

	questionId := c.Param("id")
	if _, err := strconv.Atoi(questionId); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid question ID",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	params, err := parseAndValidateParameters(c)
	if err != nil {
		return
	}

	domain := "stackoverflow.com"

	if strings.Contains(params.Sub, ".") {
		domain = params.Sub
	} else if params.Sub != "" {
		domain = fmt.Sprintf("%s.stackexchange.com", params.Sub)
	}

	soLink := fmt.Sprintf("https://%s/questions/%s/%s?answertab=%s", domain, questionId, params.QuestionTitle, params.SoSortValue)

	resp, err := fetchQuestionData(soLink)

	if resp.StatusCode() != 200 {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": fmt.Sprintf("Received a non-OK status code %d", resp.StatusCode()),
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	respBody := resp.String()

	respBodyReader := strings.NewReader(respBody)

	doc, err := goquery.NewDocumentFromReader(respBodyReader)
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Unable to parse question data",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	newFilteredQuestion, err := extractQuestionData(doc, domain)
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Failed to extract question data",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	answers, err := extractAnswersData(doc, domain)
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Failed to extract answer data",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	imagePolicy := "'self' https:"

	if c.MustGet("disable_images").(bool) {
		imagePolicy = "'self'"
	}

	c.HTML(200, "question.html", gin.H{
		"question":    newFilteredQuestion,
		"answers":     answers,
		"imagePolicy": imagePolicy,
		"theme":       c.MustGet("theme").(string),
		"currentUrl":  fmt.Sprintf("%s%s", os.Getenv("APP_URL"), c.Request.URL.Path),
		"sortValue":   params.SoSortValue,
		"domain":      domain,
	})

}

type viewQuestionInputs struct {
	QuestionID    string
	QuestionTitle string
	SoSortValue   string
	Sub           string
}

// parseAndValidateParameters consolidates the URL and query parameters into an easily-accessible struct.
func parseAndValidateParameters(c *gin.Context) (inputs viewQuestionInputs, err error) {

	questionId := c.Param("id")
	if _, err = strconv.Atoi(questionId); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid question ID",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	inputs.QuestionID = questionId

	sortValue := c.Query("sort_by")
	if sortValue == "" {
		sortValue = "votes"
	}

	soSortValue, ok := soSortValues[sortValue]
	if !ok {
		soSortValue = soSortValues["votes"]
	}

	inputs.SoSortValue = soSortValue

	sub := c.Param("sub")

	inputs.Sub = sub

	return
}

// fetchQuestionData sends the request to StackOverflow.
func fetchQuestionData(soLink string) (resp *resty.Response, err error) {
	client := resty.New()
	resp, err = client.R().Get(soLink)
	return
}

// extractQuestionData parses the HTML document and extracts question data.
func extractQuestionData(doc *goquery.Document, domain string) (question types.FilteredQuestion, err error) {
	// Extract the question title.
	questionTextParent := doc.Find("h1.fs-headline1").First()
	question.Title = strings.TrimSpace(questionTextParent.Children().First().Text())

	// Extract question tags.
	questionTags := utils.GetPostTags(doc.Find("div.post-layout").First())
	question.Tags = questionTags

	// Extract and process the question body.
	questionBodyParent := doc.Find("div.s-prose").First()
	questionBodyParentHTML, err := questionBodyParent.Html()
	if err != nil {
		return question, err
	}
	question.Body = template.HTML(utils.ProcessHTMLBody(questionBodyParentHTML))

	// Extract the shortened body description.
	shortenedBody := strings.TrimSpace(questionBodyParent.Text())
	shortenedBody = strings.ReplaceAll(shortenedBody, "\n", " ")
	if len(shortenedBody) > 50 {
		shortenedBody = shortenedBody[:50]
	}
	question.ShortenedBody = shortenedBody

	// Extract question comments.
	comments := utils.FindAndReturnComments(questionBodyParentHTML, domain, doc.Find("div.post-layout").First())
	question.Comments = comments

	// Extract question timestamp and author information.
	questionCard := doc.Find("div.postcell").First()
	extractMetadata(questionCard, &question, domain)

	return
}

// extractMetadata extracts author and timestamp information from a given selection.
func extractMetadata(selection *goquery.Selection, question *types.FilteredQuestion, domain string) {
	questionMetadata := selection.Find("div.user-info").First()
	question.Timestamp = questionMetadata.Find("span.relativetime").First().Text()

	questionAuthorURL := "https://" + domain
	questionAuthor := selection.Find("div.post-signature.owner div.user-info div.user-details a").First()
	question.AuthorName = questionAuthor.Text()
	questionAuthorURL += questionAuthor.AttrOr("href", "")
	question.AuthorURL = questionAuthorURL

	// Determine if the question has been edited and update author details accordingly.
	isQuestionEdited := selection.Find("a.js-gps-track").Text() == "edited"
	if isQuestionEdited {
		editedAuthor := questionMetadata.Find("a").Last()
		question.AuthorName = editedAuthor.Text()
		question.AuthorURL = "https://" + domain + editedAuthor.AttrOr("href", "")
	}
}

// extractAnswersData parses the HTML document and extracts answers data.
func extractAnswersData(doc *goquery.Document, domain string) ([]types.FilteredAnswer, error) {
	var answers []types.FilteredAnswer

	// Iterate over each answer block.
	doc.Find("div.answer").Each(func(i int, s *goquery.Selection) {
		var answer types.FilteredAnswer

		postLayout := s.Find("div.post-layout").First()

		// Extract upvotes.
		voteCell := postLayout.Find("div.votecell").First()
		voteCount := html.EscapeString(voteCell.Find("div.js-vote-count").Text())
		answer.Upvotes = voteCount

		// Check if the answer is accepted.
		answer.IsAccepted = s.HasClass("accepted-answer")

		// Extract answer body and process it.
		answerCell := postLayout.Find("div.answercell").First()
		answerBody := answerCell.Find("div.s-prose").First()
		answerBodyHTML, _ := answerBody.Html()

		// Process code blocks within the answer.
		processedAnswerBody := utils.ProcessHTMLBody(answerBodyHTML)
		answer.Body = template.HTML(processedAnswerBody)

		answer.Comments = utils.FindAndReturnComments(answerBodyHTML, domain, postLayout)

		// Extract author information and timestamp.
		extractAnswerAuthorInfo(s, &answer, domain)

		answers = append(answers, answer)
	})

	return answers, nil
}

// extractAnswerAuthorInfo extracts the author name, URL, and timestamp from an answer block.
// It directly mutates the answer.
func extractAnswerAuthorInfo(selection *goquery.Selection, answer *types.FilteredAnswer, domain string) {
	authorDetails := selection.Find("div.post-signature").Last()

	authorName := html.EscapeString(authorDetails.Find("div.user-details a").First().Text())
	authorURL := "https://" + domain + authorDetails.Find("div.user-details a").AttrOr("href", "")
	timestamp := html.EscapeString(authorDetails.Find("span.relativetime").Text())

	answer.AuthorName = authorName
	answer.AuthorURL = authorURL
	answer.Timestamp = timestamp
}
