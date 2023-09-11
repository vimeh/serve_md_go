package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"io"
	"net/http"
)

// Global variable to hold the sanitized and converted markdown
var sanitizedHTML []byte

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to upload file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Read the file
		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusInternalServerError)
			return
		}

		// Sanitize the content using Bluemonday
		p := bluemonday.UGCPolicy()
		sanitizedMD := p.SanitizeBytes(content)

		// Convert the Markdown to HTML
		htmlContent := markdown.ToHTML(sanitizedMD, nil, nil)

		// Store the sanitized HTML
		sanitizedHTML = htmlContent

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Render the upload form
	t, _ := template.ParseFiles("upload.html")
	t.Execute(w, nil)
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(sanitizedHTML)
}

func main() {
	http.HandleFunc("/edit", uploadHandler)
	http.HandleFunc("/", renderHandler)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
