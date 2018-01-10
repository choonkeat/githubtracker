package githubtracker

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"encoding/json"

	"github.com/pkg/errors"
)

type webhookIssue struct {
	isClosed       bool
	isOpened       bool
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	State          string    `json:"state"`
	URL            string    `json:"html_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	titleWas       *string
	bodyWas        *string
	trackerHTMLURL string
}

func (i *webhookIssue) StrippedBody() string {
	return strings.TrimSpace(bodyStripRegexpFor(i.trackerHTMLURL).ReplaceAllString(i.Body, ""))
}

func (i *webhookIssue) isChanged() bool {
	if i == nil {
		log.Printf("changed? false because issue=%#v", i)
		return false
	}

	if i.titleWas != nil {
		if old, new := strings.TrimSpace(*i.titleWas), strings.TrimSpace(i.Title); old != new {
			log.Printf("title changed! %#v -> %#v", old, new)
			return true
		}
	}

	if i.bodyWas != nil {
		bodyStrip := bodyStripRegexpFor(i.trackerHTMLURL)
		if old, new := sanitizeString(bodyStrip, *i.bodyWas), sanitizeString(bodyStrip, i.Body); old != new {
			log.Printf("body changed! %#v -> %#v", old, new)
			return true
		}
	}
	return false
}

func parseWebhookIssue(data []byte, htmlURL string) (*webhookIssue, error) {
	wh := githubWebhook{}
	if err := json.Unmarshal(data, &wh); err != nil {
		return nil, errors.Wrap(err, "unmarshal parse issue")
	}
	if wh.WebhookIssue == nil {
		return nil, nil
	}
	if strings.HasSuffix(strings.TrimSpace(wh.WebhookIssue.Title), noStorySuffix) {
		log.Printf("skip %#v in title: %#v", noStorySuffix, wh.WebhookIssue.Title)
		return nil, nil
	}

	wh.WebhookIssue.isClosed = (wh.Action == "closed")
	wh.WebhookIssue.isOpened = (wh.Action == "opened" || wh.Action == "reopened")
	wh.WebhookIssue.titleWas = wh.Changes["title"].String()
	wh.WebhookIssue.bodyWas = wh.Changes["body"].String()
	wh.WebhookIssue.trackerHTMLURL = htmlURL
	return wh.WebhookIssue, nil
}

func bodyStripRegexpFor(s string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i)^%s/%s/([\d]+)`, strings.TrimRight(s, "/"), "story/show"))
}

// ptStoryFromWebhookIssue returns nil if storyDetail is not meant to be updated
func ptStoryFromWebhookIssue(issue *webhookIssue) (*storyDetail, error) {
	if !issue.isClosed && !issue.isOpened && !issue.isChanged() {
		return nil, nil
	}

	buf := bytes.Buffer{}
	if err := bodyTemplate.Execute(&buf, issue); err != nil {
		return nil, errors.Wrapf(err, "template %#v", bodyTemplate)
	}

	strippedTitle := strings.TrimSpace(issue.Title)
	filters := []string{}

	bodyStrip := bodyStripRegexpFor(issue.trackerHTMLURL)
	log.Printf("bodyStrip=%s", bodyStrip.String())

	if res := bodyStrip.FindAllStringSubmatch(issue.Body, 1); res != nil {
		log.Printf("res=%#v", res)
		filters = append(filters, `id:"`+res[0][1]+`"`)
	}

	if issue.bodyWas != nil {
		if res := bodyStrip.FindAllStringSubmatch(*issue.bodyWas, 1); res != nil {
			log.Printf("res=%#v", res)
			filters = append(filters, `id:"`+res[0][1]+`"`)
		}
	}

	if res := bodyStrip.FindAllStringSubmatch(issue.Body, 1); res != nil {
		log.Printf("res=%#v", res)
		filters = append(filters, `id:"`+res[0][1]+`"`)
	}

	if issue.titleWas != nil {
		strippedTitleWas := strings.TrimSpace(*issue.titleWas)
		filters = append(filters, `name:"`+searchFriendly(strippedTitleWas)+`"`)
	}

	filters = append(filters, `name:"`+searchFriendly(strippedTitle)+`"`)

	story := storyDetail{
		Title:         strippedTitle,
		Body:          buf.String(),
		SearchFilters: filters,
		IsClosed:      issue.isClosed,
		IsOpened:      issue.isOpened,
	}
	return &story, nil
}

func searchFriendly(s string) string {
	return strings.Replace(s, "/", " ", -1)
}

func sanitizeString(re *regexp.Regexp, s string) string {
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

type githubWebhook struct {
	Action       string                 `json:"action"`
	WebhookIssue *webhookIssue          `json:"issue"`
	Changes      map[string]*changeFrom `json:"changes,omitempty"`
}

type changeFrom struct {
	// e.g.
	// "column_id: 1234"
	// "from": "old title string"
	From alwaysString `json:"from"`
}

func (c *changeFrom) String() *string {
	if c == nil {
		return nil
	}
	return &c.From.Value
}
