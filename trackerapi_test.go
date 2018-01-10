package githubtracker

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	testCases := []struct {
		givenStoryDetail storyDetail
		expectedJSON     string
	}{
		{
			givenStoryDetail: storyDetail{},
			expectedJSON:     `{}`,
		},
		{
			givenStoryDetail: storyDetail{
				Estimate: intptr(0),
			},
			expectedJSON: `{"estimate":0}`,
		},
		{
			givenStoryDetail: storyDetail{
				Estimate: intptr(1),
			},
			expectedJSON: `{"estimate":1}`,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			data, err := json.Marshal(tc.givenStoryDetail)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedJSON, string(data))
		})
	}
}
