package githubtracker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtStoryFromWebhookIssue(t *testing.T) {
	testCases := []struct {
		givenFile                   string
		expectedTitle, expectedBody string
		expectedSearchFilters       []string
		expectedIsDone              bool
		expectNoIssue               bool
		expectNoChange              bool
	}{
		{
			givenFile:             "testdata/github/issues.new.json",
			expectedTitle:         "users.email should have unique constraint",
			expectedBody:          "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five",
			expectedSearchFilters: []string{"name:\"users.email should have unique constraint\""},
		},
		{
			givenFile:             "testdata/github/issues.edited.json",
			expectedTitle:         "should have unique index on users.email column [Finished #12345]",
			expectedBody:          "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five",
			expectedSearchFilters: []string{"name:\"users.email should have unique constraint\"", "name:\"should have unique index on users.email column [Finished #12345]\""},
		},
		{
			givenFile:             "testdata/github/issues.edited-body.json",
			expectedTitle:         "should have unique index on users.email column[fixed #12345]",
			expectedBody:          "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five\r\n\r\n- [ ] what else?",
			expectedSearchFilters: []string{"name:\"should have unique index on users.email column[fixed #12345]\""},
		},
		{
			givenFile:      "testdata/github/issues.edited-body-add-ptlink.json",
			expectNoChange: true,
		},
		{
			givenFile:             "testdata/github/issues.edited-body-del-ptlink.json",
			expectedTitle:         "should have unique index on users.email column[fixed #12345]",
			expectedBody:          "https://github.com/user123/repo456/issues/1\r\n\r\notherwise one two three four five\r\n\r\n- [ ] what else?",
			expectedSearchFilters: []string{"name:\"should have unique index on users.email column[fixed #12345]\""},
		},
		{
			givenFile:      "testdata/github/issues.assigned.json",
			expectNoChange: true,
		},
		{
			givenFile:             "testdata/github/issues.created.json",
			expectedTitle:         "should create/update story on github issue create/update",
			expectedBody:          "https://github.com/user123/repo456/issues/2\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\r\n\r\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			expectedSearchFilters: []string{"name:\"should create update story on github issue create update\""},
		},
		{
			givenFile:             "testdata/github/issues.closed.json",
			expectedTitle:         "some story from ghe",
			expectedBody:          "https://github.com/user123/repo456/issues/8\r\n\r\n",
			expectedSearchFilters: []string{"id:\"153984041\"", "id:\"153984041\"", "name:\"some story from ghe\""},
			expectedIsDone:        true,
		},
		{
			givenFile:     "testdata/github/issues.created-with-nostory.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/issues.edited-with-nostory.json",
			expectNoIssue: true,
		},
		{
			givenFile:      "testdata/github/issues.labelled.json",
			expectNoChange: true,
		},
		{
			givenFile:      "testdata/github/issues.unassigned.json",
			expectNoChange: true,
		},
		{
			givenFile:     "testdata/github/projects.add-card.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/projects.add-column.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/projects.create.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/projects.move-card.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/projects.move-card2.json",
			expectNoIssue: true,
		},
		{
			givenFile:     "testdata/github/push.json",
			expectNoIssue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenFile, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.givenFile)
			if err != nil {
				t.Fatalf("readfile: %s", err.Error())
			}

			func() {
				if strings.Contains(tc.givenFile, "result") {
					return
				}
				var v githubWebhook
				err = json.NewDecoder(bytes.NewReader(data)).Decode(&v)
				if err != nil {
					t.Fatal(err.Error())
				}
				f, err := os.Create(tc.givenFile)
				if err != nil {
					t.Fatal(err.Error())
				}
				defer f.Close()
				enc := json.NewEncoder(f)
				enc.SetIndent("", "  ")
				err = enc.Encode(v)
				if err != nil {
					t.Fatal(err.Error())
				}
			}()

			issue, err := parseWebhookIssue(data, "https://www.pivotaltracker.com")
			if err != nil {
				t.Fatalf("ParseWebhookIssue: %s", err.Error())
			}

			if tc.expectNoIssue {
				assert.Nil(t, issue, "issue %#v", issue)
				return
			}
			if !assert.NotNil(t, issue, "issue") {
				return
			}

			ptstory, err := ptStoryFromWebhookIssue(issue)
			if err != nil {
				t.Fatalf("ptStoryFromWebhookIssue: %s", err.Error())
			}
			if tc.expectNoChange {
				assert.Nil(t, ptstory, "story %#v issue %#v", ptstory, issue)
				return
			}
			if ptstory == nil {
				t.Fatalf("ptStoryFromWebhookIssue returned unexpected nil")
			}

			assert.Equal(t, tc.expectedTitle, ptstory.Title, "expectedTitle")
			assert.Equal(t, tc.expectedBody, ptstory.Body, "expectedBody")
			assert.Equal(t, tc.expectedSearchFilters, ptstory.SearchFilters, "expectedSearchFilters")
			assert.Equal(t, tc.expectedIsDone, ptstory.IsDone, "expectedIsDone")
		})
	}
}
