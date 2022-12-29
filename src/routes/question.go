package routes

import (
	"anonymousoverflow/src/utils"
	"fmt"
	"html"
	"html/template"
	"os"
	"regexp"
	"strings"

	"anonymousoverflow/src/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var codeBlockRegex = regexp.MustCompile(`(?s)<pre><code>(.+?)<\/code><\/pre>`)
var questionCodeBlockRegex = regexp.MustCompile(`(?s)<pre class=".+"><code( class=".+")?>(.+?)</code></pre>`)

func ViewQuestion(c *gin.Context) {
	client := resty.New()

	questionId := c.Param("id")
	questionTitle := c.Param("title")

	soLink := fmt.Sprintf("https://stackoverflow.com/questions/%s/%s", questionId, questionTitle)

	resp, err := client.R().Get(soLink)
	if err != nil {
		panic(err)
	}

	respBody := resp.String()

	respBodyReader := strings.NewReader(respBody)

	doc, err := goquery.NewDocumentFromReader(respBodyReader)
	if err != nil {
		panic(err)
	}

	newFilteredQuestion := types.FilteredQuestion{}

	questionTextParent := doc.Find("h1.fs-headline1")

	questionText := questionTextParent.Children().First().Text()

	newFilteredQuestion.Title = questionText

	questionPostLayout := doc.Find("div.post-layout").First()

	questionBodyParent := doc.Find("div.s-prose")

	questionBodyParentHTML, err := questionBodyParent.Html()
	if err != nil {
		panic(err)
	}

	newFilteredQuestion.Body = template.HTML(questionBodyParentHTML)

	questionBodyText := questionBodyParent.Text()

	// remove all whitespace to create the shortened body desc
	shortenedBody := strings.TrimSpace(questionBodyText)

	// remove all newlines
	shortenedBody = strings.ReplaceAll(shortenedBody, "\n", " ")

	// get the first 50 chars
	shortenedBody = shortenedBody[:50]

	newFilteredQuestion.ShortenedBody = shortenedBody

	comments := utils.FindAndReturnComments(questionBodyParentHTML, questionPostLayout)
	newFilteredQuestion.Comments = comments

	// parse any code blocks and highlight them
	answerCodeBlocks := questionCodeBlockRegex.FindAllString(questionBodyParentHTML, -1)
	for _, codeBlock := range answerCodeBlocks {
		codeBlock = utils.StripBlockTags(codeBlock)

		// syntax highlight
		highlightedCodeBlock := utils.HighlightSyntaxViaContent(codeBlock)

		// replace the code block with the highlighted code block
		questionBodyParentHTML = strings.Replace(questionBodyParentHTML, codeBlock, highlightedCodeBlock, 1)
	}

	questionCard := doc.Find("div.postcell")

	questionMetadata := questionCard.Find("div.user-info")
	questionTimestamp := ""
	questionMetadata.Find("span.relativetime").Each(func(i int, s *goquery.Selection) {
		// get the second
		if i == 0 {
			if s.Text() != "" {
				// if it's not been edited, it means it's the first
				questionTimestamp = s.Text()
				return
			}
		}

		// otherwise it's the second element
		if i == 1 {
			questionTimestamp = s.Text()
			return
		}
	})

	newFilteredQuestion.Timestamp = questionTimestamp

	userDetails := questionMetadata.Find("div.user-details")

	questionAuthor := ""
	questionAuthorURL := ""

	userDetails.Find("a").Each(func(i int, s *goquery.Selection) {
		// get the second
		if i == 0 {
			if s.Text() != "" {
				// if it's not been edited, it means it's the first
				questionAuthor = s.Text()
				questionAuthorURL, _ = s.Attr("href")
				return
			}
		}

		// otherwise it's the second element
		if i == 1 {
			questionAuthor = s.Text()
			questionAuthorURL, _ = s.Attr("href")
			return
		}
	})

	newFilteredQuestion.AuthorName = questionAuthor
	newFilteredQuestion.AuthorURL = questionAuthorURL

	answers := []types.FilteredAnswer{}

	doc.Find("div.answer").Each(func(i int, s *goquery.Selection) {
		newFilteredAnswer := types.FilteredAnswer{}

		postLayout := s.Find("div.post-layout")
		voteCell := postLayout.Find("div.votecell")
		answerCell := postLayout.Find("div.answercell")
		answerBody := answerCell.Find("div.s-prose")
		answerBodyHTML, _ := answerBody.Html()

		voteCount := html.EscapeString(voteCell.Find("div.js-vote-count").Text())

		newFilteredAnswer.Upvotes = voteCount
		newFilteredAnswer.IsAccepted = s.HasClass("accepted-answer")

		answerFooter := s.Find("div.mt24")

		answerAuthorURL := ""
		answerAuthorName := ""
		answerTimestamp := ""

		answerFooter.Find("div.post-signature").Each(func(i int, s *goquery.Selection) {
			answerAuthorDetails := s.Find("div.user-details")

			if answerAuthorDetails.Length() == 0 {
				return
			}

			if answerAuthorDetails.Length() > 1 {
				if i == 0 {
					return
				}
			}

			answerAuthor := answerAuthorDetails.Find("a").First()

			answerAuthorURL = html.EscapeString(answerAuthor.AttrOr("href", ""))
			answerAuthorName = html.EscapeString(answerAuthor.Text())
			answerTimestamp = html.EscapeString(s.Find("span.relativetime").Text())
		})

		newFilteredAnswer.AuthorName = answerAuthorName
		newFilteredAnswer.AuthorURL = answerAuthorURL
		newFilteredAnswer.Timestamp = answerTimestamp

		// parse any code blocks and highlight them
		answerCodeBlocks := codeBlockRegex.FindAllString(answerBodyHTML, -1)
		for _, codeBlock := range answerCodeBlocks {
			codeBlock = utils.StripBlockTags(codeBlock)

			// syntax highlight
			highlightedCodeBlock := utils.HighlightSyntaxViaContent(codeBlock)

			// replace the code block with the highlighted code block
			answerBodyHTML = strings.Replace(answerBodyHTML, codeBlock, highlightedCodeBlock, 1)
		}

		comments = utils.FindAndReturnComments(answerBodyHTML, postLayout)

		newFilteredAnswer.Comments = comments
		newFilteredAnswer.Body = template.HTML(answerBodyHTML)

		answers = append(answers, newFilteredAnswer)
	})

	imagePolicy := "'self' https:"

	if c.MustGet("disable_images").(bool) {
		imagePolicy = "'self'"
	}

	c.HTML(200, "question.html", gin.H{
		"question":    newFilteredQuestion,
		"answers":     answers,
		"imagePolicy": imagePolicy,
		"theme":       c.MustGet("theme").(string),
		"currentUrl":  fmt.Sprintf("%s/questions/%s/%s", os.Getenv("APP_URL"), questionId, questionTitle),
	})

}
