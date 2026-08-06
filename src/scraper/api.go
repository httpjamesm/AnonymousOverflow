package scraper

import (
	"anonymousoverflow/src/types"
	"anonymousoverflow/src/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Owner struct {
	AccountID    int    `json:"account_id"`
	Reputation   int    `json:"reputation"`
	UserID       int    `json:"user_id"`
	UserType     string `json:"user_type"`
	ProfileImage string `json:"profile_image"`
	DisplayName  string `json:"display_name"`
	Link         string `json:"link"`
}

type QuestionItem struct {
	Tags             []string `json:"tags"`
	Owner            Owner    `json:"owner"`
	ViewCount        int      `json:"view_count"`
	Score            int      `json:"score"`
	LastActivityDate int      `json:"last_activity_date"`
	CreationDate     int      `json:"creation_date"`
	QuestionID       int      `json:"question_id"`
	ContentLicense   string   `json:"content_license"`
	Link             string   `json:"link"`
	Title            string   `json:"title"`
	Body             string   `json:"body"`
	IsAnswered       bool     `json:"is_answered"`
	AcceptedAnswerID int      `json:"accepted_answer_id"`
	AnswerCount      int      `json:"answer_count"`
}

type AnswerItem struct {
	Owner            Owner  `json:"owner"`
	ViewCount        int    `json:"view_count"`
	Score            int    `json:"score"`
	LastActivityDate int    `json:"last_activity_date"`
	CreationDate     int    `json:"creation_date"`
	QuestionID       int    `json:"question_id"`
	ContentLicense   string `json:"content_license"`
	Link             string `json:"link"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	IsAccepted       bool   `json:"is_accepted"`
	AnswerID         int    `json:"answer_id"`
}

type CommentItem struct {
	Owner          Owner  `json:"owner"`
	Score          int    `json:"score"`
	CreationDate   int    `json:"creation_date"`
	PostID         int    `json:"post_id"`
	CommentID      int    `json:"comment_id"`
	ContentLicense string `json:"content_license"`
	Body           string `json:"body"`
	Edited         bool   `json:"edited"`
}

type QuestionResponse struct {
	Results        []QuestionItem `json:"items"`
	HasMore        bool           `json:"has_more"`
	QuotaMax       int            `json:"quota_max"`
	QuotaRemaining int            `json:"quota_remaining"`
}

type AnswerResponse struct {
	Results        []AnswerItem `json:"items"`
	HasMore        bool         `json:"has_more"`
	QuotaMax       int          `json:"quota_max"`
	QuotaRemaining int          `json:"quota_remaining"`
}

type CommentResponse struct {
	Results        []CommentItem `json:"items"`
	HasMore        bool          `json:"has_more"`
	QuotaMax       int           `json:"quota_max"`
	QuotaRemaining int           `json:"quota_remaining"`
}

type ApiScraper struct{ ApiKey string }

const API_URL = "https://api.stackexchange.com/2.3"

func (s ApiScraper) GetQuestion(params ViewQuestionInputs) (types.FilteredQuestion, []types.FilteredAnswer, error) {
	client := resty.New()
	if s.ApiKey != "" {
		client.SetAuthToken(s.ApiKey)
	}

	filteredQuestion, err := getQuestionContent(client, params)
	if err != nil {
		return types.FilteredQuestion{}, []types.FilteredAnswer{}, err
	}

	questionComments, err := getComments(client, params, []string{params.QuestionID}, "questions")
	if err != nil {
		return types.FilteredQuestion{}, []types.FilteredAnswer{}, err
	}
	questionId, _ := strconv.Atoi(params.QuestionID)
	if comments, ok := questionComments[questionId]; ok {
		filteredQuestion.Comments = comments
	}

	filteredAnswers, err := getAnswers(client, params)
	if err != nil {
		return types.FilteredQuestion{}, []types.FilteredAnswer{}, err
	}
	var answerIds []string
	for _, filteredAnswer := range filteredAnswers {
		answerIds = append(answerIds, filteredAnswer.ID)
	}
	answerComments, err := getComments(client, params, answerIds, "answers")
	if err != nil {
		return types.FilteredQuestion{}, []types.FilteredAnswer{}, err
	}
	for i, filteredAnswer := range filteredAnswers {
		answerId, _ := strconv.Atoi(filteredAnswer.ID)
		if comments, ok := answerComments[answerId]; ok {
			filteredAnswers[i].Comments = comments
		}
	}

	return filteredQuestion, filteredAnswers, nil
}

func getQuestionContent(client *resty.Client, params ViewQuestionInputs) (types.FilteredQuestion, error) {

	resp, err := client.R().Get(fmt.Sprintf("%s/questions/%s?site=%s&filter=withbody", API_URL, params.QuestionID, params.Sub))
	if err != nil {
		return types.FilteredQuestion{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return types.FilteredQuestion{}, fmt.Errorf("Received a non-OK status code %d", resp.StatusCode())
	}

	var questionsResp QuestionResponse
	if err := json.Unmarshal(resp.Body(), &questionsResp); err != nil {
		return types.FilteredQuestion{}, err
	}
	question := questionsResp.Results[0]

	questionBody := template.HTML(utils.ProcessHTMLBody(question.Body))

	// Extract the shortened body description.
	shortenedBody := strings.TrimSpace(question.Body)
	shortenedBody = strings.ReplaceAll(shortenedBody, "\n", " ")
	if len(shortenedBody) > 50 {
		shortenedBody = shortenedBody[:50]
	}
	return types.FilteredQuestion{
		Title:         question.Title,
		Body:          questionBody,
		Timestamp:     time.Unix(int64(question.CreationDate), 0).Format(time.RFC1123),
		AuthorName:    question.Owner.DisplayName,
		AuthorURL:     question.Owner.Link,
		ShortenedBody: shortenedBody,
		Comments:      []types.FilteredComment{},
		Tags:          question.Tags,
	}, nil
}

func getAnswers(client *resty.Client, params ViewQuestionInputs) ([]types.FilteredAnswer, error) {
	resp, err := client.R().Get(fmt.Sprintf("%s/questions/%s/answers?site=%s&filter=withbody&sort=%s", API_URL, params.QuestionID, params.Sub, params.SortValue))
	if err != nil {
		return []types.FilteredAnswer{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return []types.FilteredAnswer{}, fmt.Errorf("Received a non-OK status code %d", resp.StatusCode())
	}

	var answersResp AnswerResponse
	if err := json.Unmarshal(resp.Body(), &answersResp); err != nil {
		return []types.FilteredAnswer{}, err
	}

	var answers []types.FilteredAnswer
	for _, answer := range answersResp.Results {
		answers = append(answers, types.FilteredAnswer{
			ID:         strconv.Itoa(answer.AnswerID),
			Upvotes:    strconv.Itoa(answer.Score),
			IsAccepted: answer.IsAccepted,
			AuthorName: answer.Owner.DisplayName,
			AuthorURL:  answer.Owner.Link,
			Timestamp:  time.Unix(int64(answer.CreationDate), 0).Format(time.RFC1123),
			Body:       template.HTML(utils.ProcessHTMLBody(answer.Body)),
			Comments:   []types.FilteredComment{},
		})
	}

	return answers, nil
}

func getComments(client *resty.Client, params ViewQuestionInputs, ids []string, idType string) (map[int][]types.FilteredComment, error) {
	delimitedIds := strings.Join(ids, ";")
	resp, err := client.R().Get(fmt.Sprintf("%s/%s/%s/comments?site=%s&filter=withbody&sort=%s", API_URL, idType, delimitedIds, params.Sub, params.SortValue))

	if err != nil {
		return map[int][]types.FilteredComment{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return map[int][]types.FilteredComment{}, fmt.Errorf("Received a non-OK status code %d", resp.StatusCode())
	}

	var commentsResp CommentResponse
	if err := json.Unmarshal(resp.Body(), &commentsResp); err != nil {
		return map[int][]types.FilteredComment{}, err
	}

	comments := map[int][]types.FilteredComment{}
	for _, comment := range commentsResp.Results {

		filteredComment := types.FilteredComment{
			Text:       template.HTML(utils.ProcessHTMLBody(comment.Body)),
			Timestamp:  time.Unix(int64(comment.CreationDate), 0).Format(time.RFC1123),
			AuthorName: comment.Owner.DisplayName,
			AuthorURL:  comment.Owner.Link,
			Upvotes:    strconv.Itoa(comment.Score),
		}

		if entries, ok := comments[comment.PostID]; ok {
			comments[comment.PostID] = append(entries, filteredComment)
		} else {
			comments[comment.PostID] = []types.FilteredComment{filteredComment}
		}
	}

	return comments, nil
}
