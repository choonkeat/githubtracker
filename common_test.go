package githubtracker

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlwaysString(t *testing.T) {
	testCases := []struct {
		givenJSON     string
		expectedValue string
	}{
		{
			givenJSON:     `"hello"`,
			expectedValue: "hello",
		},
		{
			givenJSON:     `999`,
			expectedValue: "999",
		},
		{
			givenJSON:     `-999`,
			expectedValue: "-999",
		},
		{
			givenJSON:     `-9.99`,
			expectedValue: "-9.99",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var v alwaysString
			if err := json.Unmarshal([]byte(tc.givenJSON), &v); err != nil {
				t.Fatalf(err.Error())
			}
			assert.Equal(t, tc.expectedValue, v.String())
		})
	}
}
