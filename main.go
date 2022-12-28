package main

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"

	"github.com/go-resty/resty/v2"
)

func main() {

	client := resty.New()

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "home.html", gin.H{})
	})

	type urlConversionRequest struct {
		URL string `form:"url" binding:"required"`
	}

	r.POST("/", func(c *gin.Context) {
		body := urlConversionRequest{}

		if err := c.ShouldBind(&body); err != nil {
			c.JSON(400, gin.H{"success": false, "message": "Invalid request body"})
			return
		}

		soLink := body.URL

		// validate URL
		if !strings.HasPrefix(soLink, "https://stackoverflow.com/questions/") {
			c.JSON(400, gin.H{"success": false, "message": "Invalid URL"})
			return
		}

		// redirect to the proxied thread
		c.Redirect(302, fmt.Sprintf("/questions/%s", strings.TrimPrefix(soLink, "https://stackoverflow.com/questions/")))

	})

	r.GET("/questions/:id/:title", func(c *gin.Context) {
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

		questionTextParent := doc.Find("h1.fs-headline1")

		questionText := questionTextParent.Children().First().Text()

		questionBodyParent := doc.Find("div.s-prose")

		questionBodyParentHTML, err := questionBodyParent.Html()
		if err != nil {
			panic(err)
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

		answers := []template.HTML{}

		doc.Find("div.answer").Each(func(i int, s *goquery.Selection) {
			postLayout := s.Find("div.post-layout")
			voteCell := postLayout.Find("div.votecell")
			answerCell := postLayout.Find("div.answercell")
			answerBody := answerCell.Find("div.s-prose")
			answerBodyHTML, _ := answerBody.Html()

			voteCount := voteCell.Find("div.js-vote-count").Text()

			if s.HasClass("accepted-answer") {
				// add <div class="answer-meta accepted">Accepted Answer</div> to the top of the answer
				answerBodyHTML = fmt.Sprintf(`<div class="answer-meta accepted">Accepted Answer - %s Upvotes</div>`, voteCount) + answerBodyHTML
			} else {
				// add <div class="answer-meta">%s Upvotes</div> to the top of the answer
				answerBodyHTML = fmt.Sprintf(`<div class="answer-meta">%s Upvotes</div>`, voteCount) + answerBodyHTML
			}

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

				answerAuthorURL = answerAuthor.AttrOr("href", "")
				answerAuthorName = answerAuthor.Text()
				answerTimestamp = s.Find("span.relativetime").Text()
			})

			// append <div class="answer-author">Answered %s by %s</div> to the bottom of the answer
			answerBodyHTML += fmt.Sprintf(`<div class="answer-author-parent"><div class="answer-author">Answered at %s by <a href="https://stackoverflow.com/%s" target="_blank" rel="noopener noreferrer">%s</a></div></div>`, answerTimestamp, answerAuthorURL, answerAuthorName)

			// get the timestamp and author

			answers = append(answers, template.HTML(answerBodyHTML))
		})

		c.HTML(200, "question.html", gin.H{
			"title":     questionText,
			"body":      template.HTML(questionBodyParentHTML),
			"timestamp": questionTimestamp,
			"author":    questionAuthor,
			"authorURL": questionAuthorURL,
			"answers":   answers,
		})

	})
	r.Run("localhost:8080")
}
