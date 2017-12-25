package crypto

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
)

// Server receives `password` in http post form, respond with `cipher` and `nonce`
type Server struct {
	PathPrefix string
	Secret     string
	GhAPIURL   string
	GhHTMLURL  string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(
			`<fieldset>
			  <legend>Add a "webhook" to PivotalTracker project settings</legend>
			  <form method="POST">
			    <input size="100" name="username" placeholder="github api username" required><br>
			    <input size="100" name="token" placeholder="github personal access token" required>
			    <small><a target="_blank" href="` + html.EscapeString(s.GhHTMLURL) + `/settings/tokens">from here</a></small>
			    <br>
			    <input size="100" name="api_url" value="` + html.EscapeString(s.GhAPIURL) + `" required><br>
			    <input size="100" name="repo" placeholder="username/repo" required><br>
			    <input size="100" name="github_html_url" value="` + html.EscapeString(s.GhHTMLURL) + `" required><br>
			    <input size="100" name="tracker_html_url" value="https://www.pivotaltracker.com" required><br>
			    <input size="100" name="target_path" value="` + path.Join(s.PathPrefix, "pivotaltracker") + `/" type="hidden"><br>
			    <input type="submit">
			  </form>
			</fieldset>

			<fieldset>
			  <legend>Add a "webhook" to Github repo settings</legend>
			  <form method="POST">
			    <input size="100" name="token" placeholder="pivotaltracker api token" required>
			    <small><a target="_blank" href="https://www.pivotaltracker.com/help/articles/api_token/">from here</a></small>
			    <br>
			    <input size="100" name="api_url" value="https://www.pivotaltracker.com/services/v5/projects/<xxx>" required><br>
			    <input size="100" name="html_url" value="https://www.pivotaltracker.com" required><br>
			    <input size="100" name="target_path" value="` + path.Join(s.PathPrefix, "github") + `/" type="hidden"><br>
			    <input type="submit">
			  </form>
			</fieldset>
		`,
		))
		return
	}

	plaintext := r.FormValue("token")
	ciphertext, noncetext, err := EncryptWithSecretENV(s.Secret, plaintext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	targetPath := r.FormValue("target_path")
	form := r.PostForm
	form.Set("token", ciphertext) // overwrite plain text
	form.Set("nonce", noncetext)
	form.Del("target_path")

	resultURL := targetPath + "?" + form.Encode()
	w.Header().Set("X-Result-URL", resultURL)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`<a href="%s">generated link (right click, copy)</a>`, html.EscapeString(resultURL))))
}

type contextKeyType int

var contextKey contextKeyType = 0

func ValuesFromContext(ctx context.Context) url.Values {
	if v, ok := ctx.Value(contextKey).(url.Values); ok {
		return v
	}
	return url.Values{}
}

func (s Server) RequireCipherNonce(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		cipher := values.Get("token")
		nonce := values.Get("nonce")
		password, err := DecryptWithSecretEnv(s.Secret, cipher, nonce)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		values.Set("token", password)
		ctx := context.WithValue(r.Context(), contextKey, values)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
