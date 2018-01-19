package githubtracker

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookStory(t *testing.T) {
	testCases := []struct {
		givenFile       string
		expectNoStory   bool
		expectNoIssue   bool
		expectedTitle   string
		expectedBody    *string
		expectedStoryID string
		expectedGhIssue issueDetail
	}{
		{
			givenFile:       "testdata/tracker/createresult.json",
			expectedTitle:   "should create/update github issue on pt story create/update",
			expectedBody:    ptr("https://github.com/user123/repo456/issues/4\n\ncreate me"),
			expectedStoryID: "153926444",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				id:            "4",
				Title:         "should create/update github issue on pt story create/update",
				Body:          "https://www.pivotaltracker.com/story/show/153926444\r\n\r\ncreate me",
				State:         "open",
				searchFilters: []string{"should create/update github issue on pt story create/update in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:     "testdata/tracker/move-to-current.json",
			expectNoStory: true,
		},
		{
			givenFile:     "testdata/tracker/searchresult.json",
			expectNoStory: true,
		},
		{
			givenFile:       "testdata/tracker/story.create.json",
			expectedTitle:   "should create a story on pivotal tracker",
			expectedBody:    ptr("https://github.com/user123/repo456/issues/1\n\notherwise one two three four five\n\n- [ ] what else?\ndone?"),
			expectedStoryID: "153937786",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				id:            "1",
				Title:         "should create a story on pivotal tracker",
				Body:          "https://www.pivotaltracker.com/story/show/153937786\r\n\r\notherwise one two three four five\n\n- [ ] what else?\ndone?",
				searchFilters: []string{"should create a story on pivotal tracker in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:       "testdata/tracker/story.create.blankbody.json",
			expectedTitle:   "should create a story on pivotal tracker",
			expectedBody:    ptr(""),
			expectedStoryID: "153937786",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				Title:         "should create a story on pivotal tracker",
				Body:          "https://www.pivotaltracker.com/story/show/153937786\r\n\r\n",
				searchFilters: []string{"should create a story on pivotal tracker in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:       "testdata/tracker/story.create2.json",
			expectedTitle:   "should create/update story on github issue create/update",
			expectedBody:    ptr("https://github.com/user123/repo456/issues/2\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum!"),
			expectedStoryID: "153937780",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				id:            "2",
				Title:         "should create/update story on github issue create/update",
				Body:          "https://www.pivotaltracker.com/story/show/153937780\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum!",
				searchFilters: []string{"should create/update story on github issue create/update in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:       "testdata/tracker/story_create_activity.json",
			expectedTitle:   "As a X I should be able to do Y",
			expectedBody:    ptr("Some description text lorem ipsum"),
			expectedStoryID: "153898290",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				Title:         "As a X I should be able to do Y",
				Body:          "https://www.pivotaltracker.com/story/show/153898290\r\n\r\nSome description text lorem ipsum",
				searchFilters: []string{"As a X I should be able to do Y in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:     "testdata/tracker/story_move_activity.json",
			expectNoStory: true,
		},
		{
			givenFile:       "testdata/tracker/story_update_activity.json",
			expectedTitle:   "Hey, World!",
			expectedBody:    ptr("Lorem body stuff"),
			expectedStoryID: "153973691",
			expectedGhIssue: issueDetail{
				repo:  "user123/repo456",
				Title: "Hey, World!",
				Body:  "https://www.pivotaltracker.com/story/show/153973691\r\n\r\nLorem body stuff",
				searchFilters: []string{
					"Hello world in:title is:issue repo:user123/repo456",
					"Hey, World! in:title is:issue repo:user123/repo456",
				},
			},
		},
		{
			givenFile:       "testdata/tracker/story_update_activity.bodyonly.json",
			expectedTitle:   "Hey, World!",
			expectedBody:    ptr("Lorem body stuff ONLY lah"),
			expectedStoryID: "153973691",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				Title:         "Hey, World!",
				Body:          "https://www.pivotaltracker.com/story/show/153973691\r\n\r\nLorem body stuff ONLY lah",
				searchFilters: []string{"Hey, World! in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:       "testdata/tracker/update.json",
			expectedTitle:   "As an X user, I should be able to do Y with Z",
			expectedBody:    nil,
			expectedStoryID: "153898290",
			expectedGhIssue: issueDetail{
				repo:  "user123/repo456",
				Title: "As an X user, I should be able to do Y with Z",
				searchFilters: []string{
					"As an X user, I must be able to do Y with Z in:title is:issue repo:user123/repo456",
					"As an X user, I should be able to do Y with Z in:title is:issue repo:user123/repo456",
				},
			},
		},
		{
			givenFile:     "testdata/tracker/update.points.json",
			expectNoStory: true,
		},
		{
			givenFile:     "testdata/tracker/update.started.json",
			expectNoStory: true,
		},
		{
			givenFile:     "testdata/tracker/story_create_activity.release.json",
			expectNoStory: true,
		},
		{
			givenFile:       "testdata/tracker/story_update_activity.with_ghurl.json",
			expectedTitle:   "do we use commit message linkage? or based on issue open/close?",
			expectedBody:    ptr("https://github.com/user123/repo456/issues/5\n\nesp since commit message linkage is using a large number `#1234` that in-theory overlaps with github issue numbering\n\nfunny thing?"),
			expectedStoryID: "153926863",
			expectedGhIssue: issueDetail{
				repo:          "user123/repo456",
				id:            "5",
				Title:         "do we use commit message linkage? or based on issue open/close?",
				Body:          "https://www.pivotaltracker.com/story/show/153926863\r\n\r\nesp since commit message linkage is using a large number `#1234` that in-theory overlaps with github issue numbering\n\nfunny thing?",
				searchFilters: []string{"do we use commit message linkage? or based on issue open/close? in:title is:issue repo:user123/repo456"},
			},
		},
		{
			givenFile:       "testdata/tracker/story.rejected.json",
			expectedTitle:   "how it works now??",
			expectedBody:    nil,
			expectedStoryID: "153983933",
			expectNoStory:   true,
		},
		{
			givenFile:       "testdata/tracker/story.deleted.json",
			expectedTitle:   "how it works now??",
			expectedBody:    nil,
			expectedStoryID: "153983933",
			expectNoStory:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenFile, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.givenFile)
			if err != nil {
				t.Fatal(err.Error())
			}

			story, err := parseWebhookStory(data, "https://github.com", "https://www.pivotaltracker.com")
			if err != nil {
				t.Fatal(err.Error())
			}
			if tc.expectNoStory {
				assert.Empty(t, story)
				return
			}

			if story == nil {
				t.Fatalf("unexpected story=nil for %s", tc.givenFile)
			}

			assert.Equal(t, tc.expectedTitle, story.Title, "title")
			assertStringValueEqual(t, tc.expectedBody, story.Body, "body")
			assert.Equal(t, tc.expectedStoryID, story.StoryID, "storyID")

			ghissue, err := ghIssueFromWebhookStory(*story, "user123/repo456", "https://github.com")
			if err != nil {
				t.Fatal(err.Error())
			}
			if ghissue == nil {
				t.Fatalf("unexpected ghissue=nil for %s", tc.givenFile)
			}

			assert.Equal(t, tc.expectedGhIssue, *ghissue)
		})
	}
}

func ptr(s string) *string {
	return &s
}

func assertStringValueEqual(t *testing.T, a, b *string, msgAndArgs ...interface{}) {
	if a == nil && b == nil {
		return
	}
	if a == nil || b == nil {
		t.Errorf("expect %#v but got %#v", a, b)
		return
	}
	assert.Equal(t, *a, *b, msgAndArgs...)
}
