package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

// Constants for file paths
const (
	markdownFilePath = "sanitized.md"
	htmlFilePath     = "rendered.html"
)

var pageTitle string

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to upload file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Sanitize and save the markdown content
		p := bluemonday.UGCPolicy()
		sanitizedMD := p.SanitizeReader(file)

		mdFile, err := os.Create(markdownFilePath)
		if err != nil {
			http.Error(w, "Unable to create markdown file", http.StatusInternalServerError)
			return
		}
		defer mdFile.Close()
		io.Copy(mdFile, sanitizedMD)

		// Convert markdown to HTML and save
		mdBytes, err := os.ReadFile(markdownFilePath)
		if err != nil {
			http.Error(w, "Unable to read markdown file", http.StatusInternalServerError)
			return
		}
		htmlContent := markdown.ToHTML(mdBytes, nil, nil)

		// Extract first top-level heading as the page title
		mdLines := strings.Split(string(mdBytes), "\n")
		for _, line := range mdLines {
			if strings.HasPrefix(line, "# ") {
				pageTitle = strings.TrimPrefix(line, "# ")
				break
			}
		}

		htmlFile, err := os.Create(htmlFilePath)
		if err != nil {
			http.Error(w, "Unable to create HTML file", http.StatusInternalServerError)
			return
		}
		defer htmlFile.Close()

		fmt.Println("Writing markdown to HTML file")
		io.WriteString(htmlFile, string(htmlContent))

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Render the upload form
	t, _ := template.ParseFiles("upload.html")
	t.Execute(w, nil)
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile(htmlFilePath)
	if err != nil {
		http.Error(w, "Content not available", http.StatusNotFound)
		return
	}

	w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
    <html>
    <head>
        <title>%s</title>
        <link rel="stylesheet" type="text/css" href="/static/styles.css">
    </head>
    <body>
  <div class="container">
    `, pageTitle)))

	w.Write(htmlContent)

	w.Write([]byte(`</div></body>
</html>`))
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Downloading markdown file")
	mdContent, err := os.ReadFile(markdownFilePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=sanitized.md")
	w.Header().Set("Content-Type", "text/markdown")
	w.Write(mdContent)
}

func main() {
	http.HandleFunc("/edit", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/", renderHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
