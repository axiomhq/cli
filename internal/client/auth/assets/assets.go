// Package assets contains assets for the auth package.
package assets

import (
	_ "embed"
	"html/template"
	"net/http"
)

var (
	//go:embed status.html.gotmpl
	statusHTMLTmpl string

	statusTmpl *template.Template
)

func init() {
	var err error
	if statusTmpl, err = template.New("status").Parse(statusHTMLTmpl); err != nil {
		panic(err)
	}
}

// WriteStatusPage writes the status page to the given response writer.
func WriteStatusPage(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return statusTmpl.Execute(w, nil)
}

// WriteStatusPage writes the status page to the given response writer.
func WriteStatusPageWithError(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return statusTmpl.Execute(w, map[string]any{"error": err.Error()})
}
