package githubtracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	changeTypeUpdate = "update"
	changeTypeDelete = "delete"
)

type webhookStory struct {
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	Body          *string `json:"body,omitempty"`
	StoryID       string  `json:"story_id"`
	CurrentState  string  `json:"current_state,omitempty"`
	titleWas      *string
	githubHTMLURL string
}

func parseWebhookStory(data []byte, githubHTMLURL string, trackerHTMLURL string) (*webhookStory, error) {
	var wh trackerWebhook
	if err := json.Unmarshal(data, &wh); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal")
	}

	var newTitle *string
	var newBody *string
	var newState *string
	story := webhookStory{}

	for _, c := range wh.Changes {
		c := c
		if c.Kind != "story" || c.StoryType == storyTypeRelease {
			continue
		}

		story.StoryID = fmt.Sprintf("%d", c.ID)
		story.Title = c.Name
		if c.OldValues.Name != nil {
			story.titleWas = c.OldValues.Name
		}
		if c.NewValues.Name != nil {
			newTitle = c.NewValues.Name
			story.Title = *newTitle
		}
		if c.NewValues.Description != nil {
			newBody = c.NewValues.Description
			story.Body = newBody
		}
		if c.NewValues.CurrentState != nil {
			newState = c.NewValues.CurrentState
			story.CurrentState = *newState
		}

		if c.ChangeType == changeTypeDelete {
			newState = &c.ChangeType
			story.CurrentState = *newState
		}
	}

	if newTitle != nil || newBody != nil || newState != nil {
		story.githubHTMLURL = githubHTMLURL
		story.URL = fmt.Sprintf("%s/story/show/%s", trackerHTMLURL, story.StoryID)
		return &story, nil
	}

	return nil, nil
}

type trackerWebhook struct {
	Changes []trackerChange `json:"changes,omitempty"`
}

type trackerChange struct {
	ID         int64               `json:"id"`
	ChangeType string              `json:"change_type"`
	Kind       string              `json:"kind"`
	StoryType  string              `json:"story_type,omitempty"`
	Name       string              `json:"name"`
	NewValues  trackerChangeValues `json:"new_values,omitempty"`
	OldValues  trackerChangeValues `json:"original_values,omitempty"`
}

type trackerChangeValues struct {
	Description  *string `json:"description,omitempty"`
	Name         *string `json:"name,omitempty"`
	CurrentState *string `json:"current_state,omitempty"`
}

func (s webhookStory) StrippedBody() string {
	if s.Body == nil {
		return ""
	}
	githubLinkBodyPrefix := regexp.MustCompile(fmt.Sprintf(`^%s\S+/issues/(\d+)`, s.githubHTMLURL))
	return strings.TrimSpace(githubLinkBodyPrefix.ReplaceAllString(*s.Body, ""))
}

var noStorySuffix = " [no story]"
var standardTrackerSearchScope = "in:title is:issue" // used to split and extract actual title

func ghIssueFromWebhookStory(story webhookStory, repo, githubHTMLURL string) (*issueDetail, error) {
	buf := bytes.Buffer{}
	if story.Body != nil {
		if err := bodyTemplate.Execute(&buf, story); err != nil {
			return nil, errors.Wrapf(err, "template %#v", bodyTemplate)
		}
	}

	issue := issueDetail{
		repo:  repo,
		Body:  buf.String(),
		Title: story.Title,
	}

	if story.titleWas != nil {
		issue.searchFilters = append(issue.searchFilters, fmt.Sprintf("%s %s repo:%s", *story.titleWas, standardTrackerSearchScope, repo))
	}
	issue.searchFilters = append(issue.searchFilters, fmt.Sprintf("%s %s repo:%s", story.Title, standardTrackerSearchScope, repo))

	githubLinkBodyPrefix := regexp.MustCompile(fmt.Sprintf(`^%s\S+/issues/(\d+)`, githubHTMLURL))
	if story.Body != nil {
		if res := githubLinkBodyPrefix.FindAllStringSubmatch(*story.Body, 1); res != nil {
			issue.id = res[0][1]
		}
	}

	// put after `searchFilters` since we don't want `noStorySuffix` to affect search
	switch s := story.CurrentState; s {
	case changeTypeDelete: // unclean; using invalid `delete` value for story state...
		issue.Title = issue.Title + noStorySuffix
		issue.Body = "" // don't touch it
	case storyStateAccepted:
		issue.State = "closed"
	case storyStateRejected, storyStatePlanned, storyStateStarted, storyStateUnstarted, storyStateUnscheduled:
		// if we're moving *into* any of these story state, the issue is `open`
		issue.State = "open"
	case storyStateFinished, storyStateDelivered:
		// not explicitly stating whether issue is close/open
		// e.g. a merged pull request may close an issue, marking story=finished
		//      it isn't right to re-open issue again
		// e.g. if story is marked as finished, but not accepted by product owner
		//      it isn't right to reach in and close issue too
	}

	return &issue, nil
}
