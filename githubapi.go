package githubtracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type issueDetail struct {
	Title         string `json:"title,omitempty"`
	Body          string `json:"body,omitempty"`
	State         string `json:"state,omitempty"`
	repo          string
	id            string
	searchFilters []string
}

type githubAPIClient interface {
	FindIssue(issue *issueDetail) (*githubSearchResultRow, error)
	CreateIssue(issue *issueDetail) error
	UpdateIssue(issue *issueDetail, rs *githubSearchResultRow) error
	GetIssue(issue *issueDetail, rs *githubSearchResultRow) (*githubGetResult, error)
}

type githubAPI struct {
	Username string
	Token    string
	URL      string
	Repo     string
	Client   *http.Client
}

type githubSearchResult struct {
	Items []githubSearchResultRow `json:"items"`
}

type githubSearchResultRow struct {
	Number int64  `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type githubGetResult struct {
	Body string `json:"body"`
}

func (g githubAPI) GetIssue(issue *issueDetail, rs *githubSearchResultRow) (*githubGetResult, error) {
	targetURL := g.URL + "/repos/" + issue.repo + "/issues/" + fmt.Sprintf("%d", rs.Number)
	data, err := g.perform("GET", targetURL, nil, http.StatusOK)
	if err != nil {
		return nil, errors.Wrapf(err, "GET %s", targetURL)
	}
	var v githubGetResult
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal %s", string(data))
	}

	return &v, nil
}

func (g githubAPI) FindIssue(issue *issueDetail) (*githubSearchResultRow, error) {
	for _, filter := range issue.searchFilters {
		expectedTitle := strings.Split(filter, standardTrackerSearchScope)[0]
		targetURL := g.URL + "/search/issues?q=" + url.QueryEscape(filter)
		data, err := g.perform("GET", targetURL, nil, http.StatusOK)
		if err != nil {
			return nil, errors.Wrapf(err, "GET %s %#v", targetURL, issue)
		}

		var result githubSearchResult
		if err = json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrapf(err, "json unmarshal")
		}

		expectedTitle = strings.TrimSpace(expectedTitle)
		var found *githubSearchResultRow
		for _, item := range result.Items {
			item := item
			if expectedTitle == strings.TrimSpace(item.Title) {
				fmt.Printf("found expect=%#v vs found=%#v\n", expectedTitle, item.Title)
				if found != nil {
					return nil, multipleMatchesError
				}
				found = &item
			} else {
				fmt.Printf("no match expect=%#v vs found=%#v\n", expectedTitle, item.Title)
			}
		}
		if found != nil {
			return found, nil
		}
	}

	return nil, nil
}

func (g githubAPI) perform(method, url string, body []byte, expectedStatus int) ([]byte, error) {
	fmt.Println(method, url, string(body))
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrapf(err, "%s %s %s", method, url, body)
	}
	req.SetBasicAuth(g.Username, g.Token)
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "perform %s %s %s", method, url, body)
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

	if expectedStatus != resp.StatusCode {
		return nil, errors.Errorf("wanted %d but got %d", expectedStatus, resp.StatusCode)
	}

	return data, nil
}

func (g githubAPI) CreateIssue(issue *issueDetail) error {
	targetURL := g.URL + "/repos/" + issue.repo + "/issues"
	targetJSON, err := json.Marshal(issue)
	if err != nil {
		return errors.Wrapf(err, "json marshal")
	}
	_, err = g.perform("POST", targetURL, targetJSON, http.StatusCreated)
	return err
}

func (g githubAPI) UpdateIssue(issue *issueDetail, rs *githubSearchResultRow) error {
	targetURL := g.URL + "/repos/" + issue.repo + "/issues/" + fmt.Sprintf("%d", rs.Number)
	targetJSON, err := json.Marshal(issue)
	if err != nil {
		return errors.Wrapf(err, "json marshal")
	}
	_, err = g.perform("PATCH", targetURL, targetJSON, http.StatusOK)
	return err
}

// ensure we implement the interface
var _ githubAPIClient = githubAPI{}
