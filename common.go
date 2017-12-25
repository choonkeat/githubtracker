package githubtracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
)

var multipleMatchesError = errors.Errorf("multiple matches error")
var bodyTemplate = template.Must(template.New("body").Parse("{{ .URL }}\r\n\r\n{{ .StrippedBody }}"))

// alwaysString can always decode from JSON into string value
type alwaysString struct {
	Value string
}

// UnmarshalJSON implements interface
func (s *alwaysString) UnmarshalJSON(data []byte) (err error) {
	if !bytes.HasPrefix(data, []byte(`"`)) {
		data = append([]byte(`"`), data...)
		data = append(data, []byte(`"`)...)
	}
	return json.Unmarshal(data, &s.Value)
}

// UnmarshalJSON implements interface
func (s *alwaysString) MarshalJSON() ([]byte, error) {
	if _, err := strconv.ParseInt(s.Value, 10, 64); err != nil {
		return []byte(fmt.Sprintf("%#v", s.Value)), nil
	}
	return []byte(s.Value), nil
}

func (s *alwaysString) String() string {
	return s.Value
}

type headerBody struct {
	Header http.Header
	Body   io.ReadCloser
}

func debugHeaderBody(r headerBody) ([]byte, error) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(r.Header)
	fmt.Println(buf.String())

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body")
	}
	defer r.Body.Close()

	v := map[string]interface{}{}
	if err = json.Unmarshal(data, &v); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal")
	}

	data, err = json.Marshal(v)
	if err != nil {
		return nil, errors.Wrapf(err, "json marshal")
	}
	fmt.Println(string(data))

	return data, nil
}

func ImportWebhookIssue(filename string) error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var v githubWebhook
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func ImportWebhookStory(filename string) error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var v trackerWebhook
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
