package unknown

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"golang.org/x/net/html/charset"
	"io"
	"legion-bot-v2/cheatdetect/common"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _ common.Detector = (*Detector)(nil)

type Detector struct {
	client         *http.Client
	defaultHeaders map[string]string
}

func New() *Detector {
	return &Detector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		defaultHeaders: map[string]string{
			"accept":             "*/*",
			"accept-language":    "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
			"origin":             "https://www.unknowncheats.me",
			"priority":           "u=1, i",
			"referer":            "https://www.unknowncheats.me/forum/search.php?do=process",
			"sec-ch-ua":          `"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"Linux"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-origin",
			"user-agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
			"x-requested-with":   "XMLHttpRequest",
		},
	}
}

func (d *Detector) Name() string {
	return "unknowncheats"
}

func (d *Detector) Detect(ctx context.Context, username string) ([]common.DetectedUser, error) {
	apiURL := "https://www.unknowncheats.me/forum/ajax.php?do=usersearch"

	formData := url.Values{}
	formData.Set("securitytoken", "guest")
	formData.Set("do", "usersearch")
	formData.Set("fragment", username)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range d.defaultHeaders {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	list, err := d.parseResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return pie.Map(list.Users, func(u User) common.DetectedUser {
		return common.DetectedUser{
			Username: u.Username,
			Site:     "unknowncheats.me",
		}
	}), nil
}

func (d *Detector) parseResponse(body io.Reader) (*UsersResponse, error) {
	var usersResp UsersResponse
	decoder := xml.NewDecoder(body)

	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&usersResp); err != nil {
		return nil, fmt.Errorf("failed to decode XML response: %w", err)
	}

	return &usersResp, nil
}
