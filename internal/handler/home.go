package handler

import (
	"authorization_flow_oauth/config"
	"authorization_flow_oauth/internal/utils"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gorilla/sessions"
)

type PageData struct {
	Title       string
	Message     string
	KeycloakURL string
}

func HomeHandler(config *config.Config,
	w http.ResponseWriter,
	r *http.Request,
	store *sessions.CookieStore) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Construct the path to the template
	tmplPath := filepath.Join(cwd, "..", "internal", "view", "home.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state, err := utils.GenerateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	log.Println("Saved state : ", state)
	// Store the state in the session
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Replace this URL with your actual Keycloak authentication URL
	keycloakURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/auth?client_id=%s&response_type=code&redirect_uri=%s&state=%s&scope=openid",
		config.KeycloakBaseURL, config.Realm, config.ClientID, url.QueryEscape(config.RedirectURI), url.QueryEscape(state))

	data := PageData{
		Title:       "Welcome",
		Message:     "Please log in to continue",
		KeycloakURL: keycloakURL,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
