package main

import (
	"net/http"
	"os"
	"path"

	"github.com/choonkeat/githubtracker"
	"github.com/choonkeat/githubtracker/crypto"
)

func main() {
	cryptoServer := crypto.Server{
		Secret:     os.Getenv("SECRET"),
		PathPrefix: path.Join("/", os.Getenv("UP_STAGE")),
		GhAPIURL:   "https://api.github.com",
		GhHTMLURL:  "https://github.com",
	}
	if s := os.Getenv("GITHUB_API_URL"); s != "" {
		cryptoServer.GhAPIURL = s
	}
	if s := os.Getenv("GITHUB_HTML_URL"); s != "" {
		cryptoServer.GhHTMLURL = s
	}

	http.Handle("/github/", cryptoServer.RequireCipherNonce(githubtracker.WebhookIssueHandler{}))
	http.Handle("/pivotaltracker/", cryptoServer.RequireCipherNonce(githubtracker.WebhookStoryHandler{}))
	http.Handle("/", cryptoServer)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
