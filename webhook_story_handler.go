package githubtracker

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/choonkeat/githubtracker/crypto"
	"github.com/pkg/errors"
)

type WebhookStoryHandler struct {
}

func (s WebhookStoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	client := githubAPI{
		Client:   http.DefaultClient,
		Token:    values.Get("token"),
		Username: values.Get("username"),
		URL:      values.Get("api_url"),
		Repo:     values.Get("repo"),
	}

	if err = s.handle(data, client, values.Get("repo"), values.Get("github_html_url"), values.Get("tracker_html_url")); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s WebhookStoryHandler) handle(data []byte, client githubAPIClient, repo, githubHTMLURL, trackerHTMLURL string) error {
	story, err := parseWebhookStory(data, githubHTMLURL, trackerHTMLURL)
	if err != nil {
		return errors.Wrapf(err, "json unmarshal")
	}
	if story == nil {
		log.Println("no story")
		return nil
	}
	fmt.Printf("webhook story = %#v\n", story)

	issue, err := ghIssueFromWebhookStory(*story, repo, githubHTMLURL)
	if err != nil {
		return errors.Wrapf(err, "ghIssueFromWebhookStory %s", repo)
	}
	if issue == nil {
		log.Println("no issue")
		return nil
	}
	fmt.Printf("issue %#v\n", issue)

	found, err := client.FindIssue(issue)
	if err == multipleMatchesError {
		log.Println(err.Error())
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "FindIssue %#v", issue)
	}
	fmt.Printf("found %#v\n", found)

	if found != nil {
		if strings.HasSuffix(issue.Title, noStorySuffix) {
			// get github issue and fixup the body
			founddetail, err := client.GetIssue(issue, found)
			if err != nil {
				return errors.Wrapf(err, "GetIssue %#v", found)
			}
			issue.Body = strings.TrimSpace(bodyStripRegexpFor(trackerHTMLURL).ReplaceAllString(founddetail.Body, ""))
		}
		err = client.UpdateIssue(issue, found)
		return errors.Wrapf(err, "UpdateIssue %#v", issue)
	}

	if strings.HasSuffix(issue.Title, noStorySuffix) {
		return nil // not found? don't create; we're deleting the story...
	}

	err = client.CreateIssue(issue)
	return errors.Wrapf(err, "CreateIssue %#v", issue)
}
