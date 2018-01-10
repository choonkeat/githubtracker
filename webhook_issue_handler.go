package githubtracker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/choonkeat/githubtracker/crypto"
	"github.com/pkg/errors"
)

type WebhookIssueHandler struct {
}

func (s WebhookIssueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := debugHeaderBody(headerBody{
		Header: r.Header,
		Body:   r.Body,
	})
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values := crypto.ValuesFromContext(r.Context())
	client := trackerAPI{
		Client: http.DefaultClient,
		Token:  values.Get("token"),
		URL:    values.Get("api_url"),
	}
	if err = s.handle(data, client, values.Get("html_url")); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s WebhookIssueHandler) handle(data []byte, client trackerAPIClient, htmlURL string) error {
	issue, err := parseWebhookIssue(data, htmlURL)
	if err != nil {
		return errors.Wrapf(err, "parse data")
	}
	if issue == nil {
		log.Println("no issue")
		return nil
	}

	story, err := ptStoryFromWebhookIssue(issue)
	if err != nil {
		return errors.Wrapf(err, "ptStoryFromWebhookIssue")
	}
	if story == nil {
		log.Println("no story")
		return nil
	}

	rs, err := client.FindStory(story)
	if err == multipleMatchesError {
		log.Println(err.Error()) // logging here since we're returning nil
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "FindStory %#v", story)
	}
	if rs == nil {
		if story.IsDone {
			// finishing an issue that had no story?
			// issue was created before github-pt sync
			// don't do anything on pt, let the issue close
			return nil
		}
		if err = client.CreateStory(story); err != nil {
			return errors.Wrapf(err, "CreateStory %#v", story)
		}
		return nil
	}

	if story.IsDone {
		if found, err := client.GetStory(rs.ID.String()); err == nil {
			fmt.Printf("found %#v\n", found)
			switch cs := found.CurrentState; cs {
			case storyStateStarted, storyStatePlanned, storyStateUnstarted, storyStateUnscheduled, storyStateRejected:
				if found.StoryType == storyTypeFeature && found.Estimate == 0 {
					story.CurrentState = storyStateAccepted
					story.StoryType = storyTypeChore
				} else if found.StoryType == storyTypeFeature {
					story.CurrentState = storyStateFinished
				} else {
					story.CurrentState = storyStateAccepted
				}
			}
		}
	}

	if err = client.UpdateStory(story, rs); err != nil {
		return errors.Wrapf(err, "UpdateStory %#v", story)
	}

	return nil
}
