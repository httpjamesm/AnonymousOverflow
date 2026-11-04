package routes

import (
	"anonymousoverflow/config"
	"anonymousoverflow/src/scraper"
	"anonymousoverflow/src/utils"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var codeBlockRegex = regexp.MustCompile(`(?s)<pre><code>(.+?)<\/code><\/pre>`)
var questionCodeBlockRegex = regexp.MustCompile(`(?s)<pre class=".+"><code( class=".+")?>(.+?)</code></pre>`)

func ViewQuestion(c *gin.Context) {

	questionId := c.Param("id")
	if _, err := strconv.Atoi(questionId); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid question ID",
			"version":      config.Version,
		})
		return
	}

	params, err := parseAndValidateParameters(c)
	if err != nil {
		return
	}

	var questionScraper scraper.QuestionsScraper
	if strings.ToLower(os.Getenv("SCRAPER")) == "api" {
		questionScraper = scraper.ApiScraper{ApiKey: os.Getenv("API_KEY")}
	} else {
		questionScraper = scraper.HtmlScraper{}
	}

	newFilteredQuestion, answers, err := questionScraper.GetQuestion(params)
	if err != nil {
		c.HTML(500, "home.html", gin.H{
			"errorMessage": err,
			"version":      config.Version,
		})
		return
	}

	imagePolicy := "'self' https:"

	if c.MustGet("disable_images").(bool) {
		imagePolicy = "'self'"
	}

	theme := utils.GetThemeFromEnv()

	c.HTML(200, "question.html", gin.H{
		"question":    newFilteredQuestion,
		"answers":     answers,
		"imagePolicy": imagePolicy,
		"currentUrl":  fmt.Sprintf("%s%s", os.Getenv("APP_URL"), c.Request.URL.Path),
		"sortValue":   params.SortValue,
		"domain":      params.Sub,
		"theme":       theme,
	})

}

// parseAndValidateParameters consolidates the URL and query parameters into an easily-accessible struct.
func parseAndValidateParameters(c *gin.Context) (inputs scraper.ViewQuestionInputs, err error) {

	questionId := c.Param("id")
	if _, err = strconv.Atoi(questionId); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid question ID",
			"version":      config.Version,
		})
		return
	}

	inputs.QuestionID = questionId

	sortValue := c.Query("sort_by")
	if sortValue == "" {
		sortValue = "votes"
	}
	inputs.SortValue = sortValue

	host, ok := utils.ExchangeHost(c.Param("sub"))
	if !ok {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid Stack Exchange site",
			"version":      config.Version,
		})
		err = errors.New("invalid Stack Exchange site")
		return
	}
	inputs.Sub = host

	return
}
