package githubtracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	storyStateAccepted    = "accepted"
	storyStateDelivered   = "delivered"
	storyStateFinished    = "finished"
	storyStateStarted     = "started"
	storyStateRejected    = "rejected"
	storyStatePlanned     = "planned"
	storyStateUnstarted   = "unstarted"
	storyStateUnscheduled = "unscheduled"

	storyTypeFeature = "feature"
	storyTypeBug     = "bug"
	storyTypeChore   = "chore"
	storyTypeRelease = "release"
)

// payload to be sent to pivotaltracker.com
type storyDetail struct {
	Title         string   `json:"name,omitempty"`
	Body          string   `json:"description,omitempty"`
	SearchFilters []string `json:"-"`
	IsDone        bool     `json:"-"`
	Estimate      *int     `json:"estimate,omitempty"`
	CurrentState  string   `json:"current_state,omitempty"`
	StoryType     string   `json:"story_type,omitempty"`
}

type trackerAPIClient interface {
	FindStory(story *storyDetail) (*trackerSearchResultRow, error)
	CreateStory(story *storyDetail) error
	UpdateStory(story *storyDetail, rs *trackerSearchResultRow) error
	GetStory(storyID string) (*trackerSearchResultRow, error)
	RequiresChoreEstimate() bool
}

type trackerAPI struct {
	Token          string
	URL            string
	HTMLURL        string
	EstimateChores bool
	Client         *http.Client
}

type trackerSearchResult struct {
	Stories struct {
		Stories []trackerSearchResultRow `json:"stories"`
	} `json:"stories"`
}

type trackerSearchResultRow struct {
	ID           alwaysString
	Name         string
	Description  string
	Estimate     int
	Kind         string
	StoryType    string `json:"story_type"`
	CurrentState string `json:"current_state"`
}

var titleInSearch = regexp.MustCompile(`^name:"(.+)"$`)

func (t trackerAPI) FindStory(story *storyDetail) (*trackerSearchResultRow, error) {
	for _, filter := range story.SearchFilters {
		expectedTitle := story.Title
		if result := titleInSearch.FindAllStringSubmatch(filter, 1); result != nil {
			expectedTitle = result[0][1]
		}

		// drop special characters from `filter` (pt search cannot handle)
		targetURL := t.URL + "/search?query=" + url.QueryEscape(filter)
		data, err := t.perform("GET", targetURL, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "GET %s %#v", targetURL, story)
		}

		result := trackerSearchResult{}
		if err = json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrapf(err, "json unmarshal")
		}

		// if too many matches, abort?; potential infinite loop
		expectedTitle = strings.TrimSpace(expectedTitle)
		var found *trackerSearchResultRow
		for _, item := range result.Stories.Stories {
			item := item
			if expectedTitle == strings.TrimSpace(item.Name) {
				fmt.Printf("found expect=%#v vs found=%#v\n", expectedTitle, item.Name)
				if found != nil {
					return nil, multipleMatchesError
				}
				found = &item
			} else {
				fmt.Printf("no match expect=%#v vs found=%#v\n", expectedTitle, item.Name)
			}
		}
		if found != nil {
			return found, nil
		}
	}
	return nil, nil
}

func (t trackerAPI) perform(method, url string, body []byte) ([]byte, error) {
	fmt.Println(method, url, string(body))
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrapf(err, "new request: %s %s %s", method, url, string(body))
	}
	req.Header.Set("X-TrackerToken", t.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "client do: %s %s %s", method, url, string(body))
	}

	// start debug
	data, err := debugHeaderBody(headerBody{
		Header: resp.Header,
		Body:   resp.Body,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "debug headerbody")
	}
	// end debug

	return data, nil
}

func (t trackerAPI) CreateStory(story *storyDetail) error {
	targetURL := t.URL + "/stories"
	targetJSON, err := json.Marshal(story)
	if err != nil {
		return errors.Wrapf(err, "json marshal")
	}
	_, err = t.perform("POST", targetURL, targetJSON)
	return err
}

func (t trackerAPI) GetStory(storyID string) (*trackerSearchResultRow, error) {
	targetURL := t.URL + "/stories/" + storyID
	data, err := t.perform("GET", targetURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "GET %s", targetURL)
	}
	rs := trackerSearchResultRow{}
	if err = json.Unmarshal(data, &rs); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal")
	}
	return &rs, nil
}

func (t trackerAPI) UpdateStory(story *storyDetail, rs *trackerSearchResultRow) error {
	targetURL := t.URL + "/stories/" + rs.ID.String()
	targetJSON, err := json.Marshal(story)
	if err != nil {
		return errors.Wrapf(err, "json marshal")
	}

	_, err = t.perform("PUT", targetURL, targetJSON)
	return err
}

func (t trackerAPI) RequiresChoreEstimate() bool {
	return t.EstimateChores
}

// ensure we implement the interface
var _ trackerAPIClient = trackerAPI{}
