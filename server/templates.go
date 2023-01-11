package server

import (
	"encoding/gob"
	"github.com/gork74/aws-parameter-bulk/pkg/forms"
	"github.com/gork74/aws-parameter-bulk/pkg/models"
	"github.com/gork74/aws-parameter-bulk/ui"
	"html/template"
	"io/fs"
	"path/filepath"
)

// Add templateData to gob to make it serializable
func InitTemplates() {
	gob.RegisterName("github.com/gork74/aws-parameter-bulk/server.templateData", &templateData{})
}

type templateData struct {
	NamesLeft      string
	JsonLeft       bool
	RecursiveLeft  bool
	NamesRight     string
	JsonRight      bool
	RecursiveRight bool
	Different      bool
	CSRFToken      string
	Flash          string
	FlashError     string
	Form           *forms.Form
	Compare        []models.ValueCompare
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{}

func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the fs.Glob function to get a slice of all filepaths with
	// the extension '.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := fs.Glob(ui.Files, "html/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		patterns := []string{
			"html/*.layout.tmpl",
			"html/*.page.tmpl",
			"html/*.partial.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}

	// Return the map.
	return cache, nil
}
