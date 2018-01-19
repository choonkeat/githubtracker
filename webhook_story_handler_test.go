package githubtracker

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type logGithubClient struct {
	History            []logAction
	ExpectedFoundIssue *githubSearchResultRow
	ExpectedError      error
}

type logAction struct {
	Method             string
	GivenID            string
	GivenTitle         string
	GivenBody          string
	GivenState         string
	GivenSearchFilters []string
}

func (l *logGithubClient) GetIssue(issue *issueDetail, rs *githubSearchResultRow) (*githubGetResult, error) {
	l.History = append(l.History, logAction{
		Method:  "GetIssue",
		GivenID: fmt.Sprintf("%d", rs.Number),
	})
	if l.ExpectedFoundIssue == nil {
		return nil, l.ExpectedError
	}
	return &githubGetResult{Body: l.ExpectedFoundIssue.Body}, nil
}

func (l *logGithubClient) FindIssue(issue *issueDetail) (*githubSearchResultRow, error) {
	l.History = append(l.History, logAction{
		Method:             "FindIssue",
		GivenID:            issue.id,
		GivenTitle:         issue.Title,
		GivenBody:          issue.Body,
		GivenState:         issue.State,
		GivenSearchFilters: issue.searchFilters,
	})
	return l.ExpectedFoundIssue, l.ExpectedError
}

func (l *logGithubClient) CreateIssue(issue *issueDetail) error {
	l.History = append(l.History, logAction{
		Method:     "CreateIssue",
		GivenID:    issue.id,
		GivenTitle: issue.Title,
		GivenBody:  issue.Body,
		GivenState: issue.State,
	})
	return l.ExpectedError
}

func (l *logGithubClient) UpdateIssue(issue *issueDetail, rs *githubSearchResultRow) error {
	l.History = append(l.History, logAction{
		Method:     "UpdateIssue",
		GivenID:    fmt.Sprintf("%d", rs.Number),
		GivenTitle: issue.Title,
		GivenBody:  issue.Body,
		GivenState: issue.State,
	})
	return l.ExpectedError
}

func TestGithubAPIClient(t *testing.T) {
	testCases := []struct {
		givenFile       string
		givenFoundIssue *githubSearchResultRow
		givenError      error
		expectedHistory []logAction
	}{
		{
			givenFile: "testdata/tracker/createresult.json",
			expectedHistory: []logAction{
				{
					Method:             "FindIssue",
					GivenID:            "4",
					GivenTitle:         "should create/update github issue on pt story create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153926444\r\n\r\ncreate me",
					GivenState:         "open",
					GivenSearchFilters: []string{"should create/update github issue on pt story create/update in:title is:issue repo:user123/repo456"},
				},
				{
					Method:             "CreateIssue",
					GivenID:            "4",
					GivenTitle:         "should create/update github issue on pt story create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153926444\r\n\r\ncreate me",
					GivenState:         "open",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/move-to-current.json",
		},
		{
			givenFile: "testdata/tracker/searchresult.json",
		},
		{
			givenFile: "testdata/tracker/story.create.json",
			expectedHistory: []logAction{
				{
					Method:             "FindIssue",
					GivenID:            "1",
					GivenTitle:         "should create a story on pivotal tracker",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153937786\r\n\r\notherwise one two three four five\n\n- [ ] what else?\ndone?",
					GivenSearchFilters: []string{"should create a story on pivotal tracker in:title is:issue repo:user123/repo456"},
				},
				{
					Method:             "CreateIssue",
					GivenID:            "1",
					GivenTitle:         "should create a story on pivotal tracker",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153937786\r\n\r\notherwise one two three four five\n\n- [ ] what else?\ndone?",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/story.create2.json",
			expectedHistory: []logAction{
				{
					Method:             "FindIssue",
					GivenID:            "2",
					GivenTitle:         "should create/update story on github issue create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153937780\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum!",
					GivenSearchFilters: []string{"should create/update story on github issue create/update in:title is:issue repo:user123/repo456"},
				},
				{
					Method:             "CreateIssue",
					GivenID:            "2",
					GivenTitle:         "should create/update story on github issue create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153937780\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum!",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/story_create_activity.json",
			expectedHistory: []logAction{
				{
					Method:             "FindIssue",
					GivenID:            "",
					GivenTitle:         "As a X I should be able to do Y",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153898290\r\n\r\nSome description text lorem ipsum",
					GivenSearchFilters: []string{"As a X I should be able to do Y in:title is:issue repo:user123/repo456"},
				},
				{
					Method:             "CreateIssue",
					GivenID:            "",
					GivenTitle:         "As a X I should be able to do Y",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153898290\r\n\r\nSome description text lorem ipsum",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/story_move_activity.json",
		},
		{
			givenFile: "testdata/tracker/update.json",
			givenFoundIssue: &githubSearchResultRow{
				Number: 42,
				Title:  "As an X user, I should be able to do Y with Z",
				Body:   "Hello world",
			},
			expectedHistory: []logAction{
				{
					Method:     "FindIssue",
					GivenID:    "",
					GivenTitle: "As an X user, I should be able to do Y with Z",
					GivenSearchFilters: []string{
						"As an X user, I must be able to do Y with Z in:title is:issue repo:user123/repo456",
						"As an X user, I should be able to do Y with Z in:title is:issue repo:user123/repo456",
					},
				},
				{
					Method:             "UpdateIssue",
					GivenID:            "42",
					GivenTitle:         "As an X user, I should be able to do Y with Z",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/update.points.json",
		},
		{
			givenFile: "testdata/tracker/update.started.json",
		},
		{
			givenFile: "testdata/tracker/updateresult.json",
		},
		{
			givenFile: "testdata/tracker/story_update_activity2.json",
			expectedHistory: []logAction{
				{
					Method:             "FindIssue",
					GivenID:            "4",
					GivenTitle:         "should create/update github issue on pt story create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153926473\r\n\r\nok last bit",
					GivenSearchFilters: []string{"should create/update github issue on pt story create/update in:title is:issue repo:user123/repo456"},
				},
				{
					Method:             "CreateIssue",
					GivenID:            "4",
					GivenTitle:         "should create/update github issue on pt story create/update",
					GivenBody:          "https://www.pivotaltracker.com/story/show/153926473\r\n\r\nok last bit",
					GivenSearchFilters: []string(nil),
				},
			},
		},
		{
			givenFile: "testdata/tracker/story_update_activity.accepted.json",
			givenFoundIssue: &githubSearchResultRow{
				Number: 42,
				Title:  "should create/update github issue on pt story create/update",
				Body:   "Hello world",
			},
			expectedHistory: []logAction(nil),
		},
		{
			givenFile: "testdata/tracker/story_update_activity.delete.json",
			givenFoundIssue: &githubSearchResultRow{
				Number: 42,
				Title:  "should create/update github issue on pt story create/update",
				Body:   "https://www.pivotaltracker.com/story/show/153926473\r\n\r\nHello world",
			},
			expectedHistory: []logAction(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenFile, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.givenFile)
			if err != nil {
				t.Fatalf("readfile: %s", err.Error())
			}

			logclient := logGithubClient{
				ExpectedFoundIssue: tc.givenFoundIssue,
				ExpectedError:      tc.givenError,
			}
			s := WebhookStoryHandler{}
			err = s.handle(data, &logclient, "user123/repo456", "https://github.com", "https://www.pivotaltracker.com")
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedHistory, logclient.History)
		})
	}
}
