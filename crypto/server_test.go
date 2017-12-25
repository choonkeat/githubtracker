package crypto

import (
	"bytes"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	testCases := []struct {
		givenFormValues url.Values
		expectedStatus  int
	}{
		{
			givenFormValues: url.Values{"token": []string{"h3llo+w0rl!"}},
			expectedStatus:  http.StatusOK,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := httptest.NewRequest("POST", "http://example.com/abc/", bytes.NewReader([]byte(tc.givenFormValues.Encode())))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			s := Server{
				Secret: uuid.New().String(),
			}
			s.ServeHTTP(w, r)
			result := w.Result()

			assert.Equal(t, tc.expectedStatus, result.StatusCode, "http status")
			data, err := ioutil.ReadAll(result.Body)
			defer result.Body.Close()
			assert.Nil(t, err, "read body")

			resultURL, err := url.Parse(result.Header.Get("X-Result-URL"))
			assert.Nil(t, err, "x-result-url parse")
			assert.Contains(t, string(data), html.EscapeString(resultURL.String()))

			token := resultURL.Query().Get("token")
			nonce := resultURL.Query().Get("nonce")
			assert.NotEmpty(t, token, "cipher text")
			assert.NotEmpty(t, nonce, "nonce text")

			plaintext, err := DecryptWithSecretEnv(s.Secret, token, nonce)
			assert.Nil(t, err, "decrypt")
			assert.Equal(t, tc.givenFormValues.Get("token"), plaintext)
		})
	}
}
