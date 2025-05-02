package steam_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"legion-bot-v2/api/dao"
	"legion-bot-v2/util/taskq"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrRequestFailed     = errors.New("request to Steam API failed")
	ErrInvalidStatusCode = errors.New("invalid status code from Steam API")
	ErrAPIRequestFailure = errors.New("Steam API reported failure")
	ErrNoCommentsFound   = errors.New("no comments found in response")
	ErrHTMLParsingFailed = errors.New("failed to parse HTML comments")
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

type Client struct {
	client           *http.Client
	sessionId        string
	steamSecureLogin string
	queue            *taskq.Queue
}

func NewClient(sessionId, steamSecureLogin string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar: %v", err)
	}

	steamCommunityURL, _ := url.Parse("https://steamcommunity.com")
	cookies := []*http.Cookie{
		{Name: "sessionid", Value: sessionId},
		{Name: "steamLoginSecure", Value: steamSecureLogin},
	}
	jar.SetCookies(steamCommunityURL, cookies)

	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
		sessionId:        sessionId,
		steamSecureLogin: steamSecureLogin,
		queue:            taskq.New(1, 1.0/3.0, 1), // 1 request per 3 seconds
	}, nil
}

func (c *Client) GetLatestComments(steamID string) ([]dao.Comment, error) {
	return taskq.ComputeWithError(c.queue, func() ([]dao.Comment, error) {
		url := fmt.Sprintf("https://steamcommunity.com/comment/Profile/render/%s/-1/", steamID)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
		}

		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%w: %d", ErrInvalidStatusCode, resp.StatusCode)
		}

		var apiResponse CommentsResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
		}

		if !apiResponse.Success {
			return nil, ErrAPIRequestFailure
		}

		if apiResponse.CommentsHTML == "" {
			return nil, ErrNoCommentsFound
		}

		comments, err := parseCommentsHTML(apiResponse.CommentsHTML)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrHTMLParsingFailed, err)
		}

		return comments, nil
	})
}

func parseCommentsHTML(html string) ([]dao.Comment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var comments []dao.Comment

	doc.Find(".commentthread_comment").Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		author := strings.TrimSpace(s.Find(".commentthread_author_link bdi").Text())
		text := strings.TrimSpace(s.Find(".commentthread_comment_text").Text())

		timestampStr, exists := s.Find(".commentthread_comment_timestamp").Attr("data-timestamp")
		var timestamp time.Time
		if exists {
			var ts int64
			if _, err := fmt.Sscanf(timestampStr, "%d", &ts); err == nil {
				timestamp = time.Unix(ts, 0)
			}
		}

		if author != "" && text != "" {
			comments = append(comments, dao.Comment{
				ID:        strings.TrimPrefix(id, "comment_"),
				Author:    author,
				Text:      text,
				Timestamp: timestamp,
			})
		}
	})

	if len(comments) == 0 {
		return nil, ErrNoCommentsFound
	}

	return comments, nil
}

func (c *Client) DeleteComment(steamID string, commentID string) error {
	return taskq.Compute(c.queue, func() error {
		formData := url.Values{
			"sessionid":  {c.sessionId},
			"gidcomment": {commentID},
			"start":      {"0"},
			"count":      {"6"},
			"feature2":   {"-1"},
		}

		req, err := http.NewRequest("POST", "https://steamcommunity.com/comment/Profile/delete/"+steamID+"/", bytes.NewBufferString(formData.Encode()))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", "https://steamcommunity.com/profiles/"+steamID+"/")
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to delete comment, status code: %d", resp.StatusCode)
		}

		return nil
	})
}

func (c *Client) PostComment(steamID, comment string) (string, error) {
	return taskq.ComputeWithError(c.queue, func() (string, error) {
		formData := url.Values{
			"sessionid":   {c.sessionId},
			"comment":     {comment},
			"feature2":    {"-1"},
			"count":       {"6"},
			"publishedfp": {"0"},
		}

		req, err := http.NewRequest("POST", "https://steamcommunity.com/comment/Profile/post/"+steamID+"/-1/", bytes.NewBufferString(formData.Encode()))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", "https://steamcommunity.com/profiles/"+steamID+"/")
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to post comment, status code: %d", resp.StatusCode)
		}

		// Parse the response to get the comment ID
		var response struct {
			Success      bool   `json:"success"`
			CommentsHTML string `json:"comments_html"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return "", fmt.Errorf("failed to decode response: %v", err)
		}

		if !response.Success {
			return "", fmt.Errorf("steam returned unsuccessful response")
		}

		// Parse the HTML to extract the comment ID
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.CommentsHTML))
		if err != nil {
			return "", fmt.Errorf("failed to parse comments HTML: %v", err)
		}

		commentDiv := doc.Find(".commentthread_comment").First()
		if commentDiv.Length() == 0 {
			return "", fmt.Errorf("no comment found in response")
		}

		commentID, exists := commentDiv.Attr("id")
		if !exists {
			return "", fmt.Errorf("comment ID not found in response")
		}

		return strings.TrimPrefix(commentID, "comment_"), nil
	})
}
