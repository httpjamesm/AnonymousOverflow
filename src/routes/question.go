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
	client := resty.New()

	questionId := c.Param("id")
	if _, err := strconv.Atoi(questionId); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid question ID",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	questionTitle := c.Param("title")

	sortValue := c.Query("sort_by")
	if sortValue == "" {
		sortValue = "votes"
	}

	soSortValue, ok := soSortValues[sortValue]
	if !ok {
		soSortValue = soSortValues["votes"]
	}

	sub := c.Param("sub")

	domain := "stackoverflow.com"

	if sub != "" {
		domain = fmt.Sprintf("%s.stackexchange.com", sub)
	}

	soLink := fmt.Sprintf("https://%s/questions/%s/%s?answertab=%s", domain, questionId, questionTitle, soSortValue)

	resp, err := client.R().Get(soLink)
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Unable to fetch question data",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}
	defer resp.RawResponse.Body.Close()

	if resp.StatusCode() != 200 {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Received a non-OK status code",
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

	newFilteredQuestion := types.FilteredQuestion{}

	questionTextParent := doc.Find("h1.fs-headline1")

	questionText := questionTextParent.Children().First().Text()

	newFilteredQuestion.Title = questionText

	questionPostLayout := doc.Find("div.post-layout").First()

	questionTags := utils.GetPostTags(questionPostLayout)
	newFilteredQuestion.Tags = questionTags

	questionBodyParent := doc.Find("div.s-prose")

	questionBodyParentHTML, err := questionBodyParent.Html()
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": "Unable to parse question body",
			"theme":        c.MustGet("theme").(string),
			"version":      config.Version,
		})
		return
	}

	newFilteredQuestion.Body = template.HTML(utils.ReplaceImgTags(questionBodyParentHTML))

	questionBodyText := questionBodyParent.Text()

	// remove all whitespace to create the shortened body desc
	shortenedBody := strings.TrimSpace(questionBodyText)

	// remove all newlines
	shortenedBody = strings.ReplaceAll(shortenedBody, "\n", " ")

	// get the first 50 chars
	shortenedBody = shortenedBody[:50]

	newFilteredQuestion.ShortenedBody = shortenedBody

	comments := utils.FindAndReturnComments(questionBodyParentHTML, domain, questionPostLayout)
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
	questionAuthorURL := fmt.Sprintf("https://%s", domain)

	// check if the question has been edited
	isQuestionEdited := false
	questionMetadata.Find("a.js-gps-track").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "edited") {
			isQuestionEdited = true
			return
		}
	})

	userDetails.Find("a").Each(func(i int, s *goquery.Selection) {
		// if question has been edited, the author is the second element.

		if isQuestionEdited {
			if i == 1 {
				questionAuthor = s.Text()
				questionAuthorURL += s.AttrOr("href", "")
				return
			}
		} else {
			// otherwise it's the first element
			if i == 0 {
				questionAuthor = s.Text()
				questionAuthorURL += s.AttrOr("href", "")
				return
			}
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

		answerAuthorURL := fmt.Sprintf("https://%s", domain)
		answerAuthorName := ""
		answerTimestamp := ""

		answerFooter.Find("div.post-signature").Each(func(i int, s *goquery.Selection) {
			answerAuthorDetails := s.Find("div.user-details")

			if answerAuthorDetails.Length() > 0 {
				questionAuthor := answerAuthorDetails.Find("a").First()
				answerAuthorName = html.EscapeString(questionAuthor.Text())
				answerAuthorURL += html.EscapeString(questionAuthor.AttrOr("href", ""))
			}

			answerTimestamp = html.EscapeString(s.Find("span.relativetime").Text())
		})

		answerId, _ := s.Attr("data-answerid")
		newFilteredAnswer.ID = answerId

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

		comments = utils.FindAndReturnComments(answerBodyHTML, domain, postLayout)

		newFilteredAnswer.Comments = comments
		newFilteredAnswer.Body = template.HTML(utils.ReplaceImgTags(answerBodyHTML))

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
		"sortValue":   sortValue,
		"domain":      domain,
	})

}
