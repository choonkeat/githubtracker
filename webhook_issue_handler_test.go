package githubtracker

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type logTrackerClient struct {
	History            []logTrackerAction
	ExpectedFoundStory *trackerSearchResultRow
	ExpectedError      error
	EstimateChores     bool
}

type logTrackerAction struct {
	Method             string
	GivenID            string
	GivenTitle         string
	GivenBody          string
	GivenIsClosed      bool
	GivenSearchFilters []string
	GivenEstimate      *int
	GivenCurrentState  string
	GivenStoryType     string
}

func (l *logTrackerClient) GetStory(storyID string) (*trackerSearchResultRow, error) {
	l.History = append(l.History, logTrackerAction{
		Method:  "GetStory",
		GivenID: storyID,
	})
	if l.ExpectedFoundStory == nil {
		return nil, l.ExpectedError
	}
	return l.ExpectedFoundStory, nil
}

func (l *logTrackerClient) FindStory(story *storyDetail) (*trackerSearchResultRow, error) {
	l.History = append(l.History, logTrackerAction{
		Method:             "FindStory",
		GivenTitle:         story.Title,
		GivenBody:          story.Body,
		GivenIsClosed:      story.IsClosed,
		GivenSearchFilters: story.SearchFilters,
		GivenEstimate:      story.Estimate,
		GivenCurrentState:  story.CurrentState,
		GivenStoryType:     story.StoryType,
	})
	return l.ExpectedFoundStory, l.ExpectedError
}

func (l *logTrackerClient) CreateStory(story *storyDetail) error {
	l.History = append(l.History, logTrackerAction{
		Method:             "CreateIssue",
		GivenTitle:         story.Title,
		GivenBody:          story.Body,
		GivenIsClosed:      story.IsClosed,
		GivenSearchFilters: story.SearchFilters,
		GivenEstimate:      story.Estimate,
		GivenCurrentState:  story.CurrentState,
		GivenStoryType:     story.StoryType,
	})
	return l.ExpectedError
}

func (l *logTrackerClient) UpdateStory(story *storyDetail, rs *trackerSearchResultRow) error {
	l.History = append(l.History, logTrackerAction{
		Method:             "UpdateStory",
		GivenID:            rs.ID.String(),
		GivenTitle:         story.Title,
		GivenBody:          story.Body,
		GivenIsClosed:      story.IsClosed,
		GivenSearchFilters: story.SearchFilters,
		GivenEstimate:      story.Estimate,
		GivenCurrentState:  story.CurrentState,
		GivenStoryType:     story.StoryType,
	})
	return l.ExpectedError
}

func (l *logTrackerClient) RequiresChoreEstimate() bool {
	return l.EstimateChores
}

func intptr(i int) *int {
	return &i
}

func TestTrackerAPIClient(t *testing.T) {
	testCases := []struct {
		givenFile                 string
		givenFoundStory           *trackerSearchResultRow
		givenError                error
		givenChoresCanBeEstimated bool
		expectedHistory           []logTrackerAction
	}{
		{
			givenFile: "testdata/github/issues.new.json",
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "users.email should have unique constraint", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"users.email should have unique constraint\""}},
				logTrackerAction{Method: "CreateIssue", GivenID: "", GivenTitle: "users.email should have unique constraint", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"users.email should have unique constraint\""}},
			},
		},
		{
			givenFile: "testdata/github/issues.edited.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    "feature",
				CurrentState: storyStateUnscheduled,
			},
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "should have unique index on users.email column [Finished #12345]", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"users.email should have unique constraint\"", "name:\"should have unique index on users.email column [Finished #12345]\""}},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "should have unique index on users.email column [Finished #12345]", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"users.email should have unique constraint\"", "name:\"should have unique index on users.email column [Finished #12345]\""}},
			},
		},
		{
			givenFile: "testdata/github/issues.edited-body.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    "feature",
				CurrentState: storyStateUnscheduled,
			},
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "should have unique index on users.email column[fixed #12345]", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five\r\n\r\n- [ ] what else?", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"should have unique index on users.email column[fixed #12345]\""}},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "should have unique index on users.email column[fixed #12345]", GivenBody: "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five\r\n\r\n- [ ] what else?", GivenIsClosed: false, GivenSearchFilters: []string{"name:\"should have unique index on users.email column[fixed #12345]\""}},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    "feature",
				CurrentState: storyStateUnscheduled,
			},
			givenChoresCanBeEstimated: true,
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenEstimate: intptr(0), GivenCurrentState: "accepted", GivenStoryType: "chore"},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    "feature",
				CurrentState: storyStateUnscheduled,
			},
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenCurrentState: "accepted", GivenStoryType: "chore"},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    storyTypeBug,
				CurrentState: storyStateUnscheduled,
			},
			givenChoresCanBeEstimated: true,
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenEstimate: intptr(0), GivenCurrentState: "accepted"},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     0,
				Kind:         "story",
				StoryType:    storyTypeBug,
				CurrentState: storyStateUnscheduled,
			},
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenCurrentState: "accepted"},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     1,
				Kind:         "story",
				StoryType:    storyTypeBug,
				CurrentState: storyStateRejected,
			},
			givenChoresCanBeEstimated: true,
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenEstimate: intptr(1), GivenCurrentState: storyStateAccepted},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			givenFoundStory: &trackerSearchResultRow{
				ID:           alwaysString{Value: "42"},
				Name:         "found title",
				Description:  "found description",
				Estimate:     1,
				Kind:         "story",
				StoryType:    storyTypeBug,
				CurrentState: storyStateRejected,
			},
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
				logTrackerAction{Method: "GetStory", GivenID: "42", GivenTitle: "", GivenBody: "", GivenIsClosed: false, GivenSearchFilters: []string(nil), GivenCurrentState: "", GivenStoryType: ""},
				logTrackerAction{Method: "UpdateStory", GivenID: "42", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}, GivenCurrentState: storyStateAccepted},
			},
		},
		{
			givenFile: "testdata/github/issues.closed.json",
			expectedHistory: []logTrackerAction{
				logTrackerAction{Method: "FindStory", GivenID: "", GivenTitle: "some story from ghe", GivenBody: "https://github.com/user123/repo456/issues/8\r\n\r\n", GivenIsClosed: true, GivenSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenFile, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.givenFile)
			if err != nil {
				t.Fatalf("readfile: %s", err.Error())
			}

			logclient := logTrackerClient{
				ExpectedFoundStory: tc.givenFoundStory,
				ExpectedError:      tc.givenError,
				EstimateChores:     tc.givenChoresCanBeEstimated,
			}
			s := WebhookIssueHandler{}
			err = s.handle(data, &logclient, "https://www.pivotaltracker.com")
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedHistory, logclient.History)
		})
	}
}
