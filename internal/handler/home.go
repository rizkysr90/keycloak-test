package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

type PageData struct {
	Title       string
	Message     string
	KeycloakURL string
}

func HomeHandler(
	w http.ResponseWriter,
	r *http.Request,
	) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
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
	data := PageData{
		Title:       "Welcome",
		Message:     "Please log in to continue",
		KeycloakURL: "http://localhost:8081/login-keycloak",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
