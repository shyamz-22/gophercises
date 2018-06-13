package main

import (
	"net/http"
	"log"
	"github.com/shyamz-22/gophercises/quoki/page"
	"html/template"
	"strings"
	"regexp"
	"github.com/pkg/errors"
	"fmt"
)

var (
	templates     = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html"))
	validRestPath = regexp.MustCompile("^/(edit|save|view)/.*")
	validPath     = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := page.LoadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusNotFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := page.LoadPage(title)
	if err != nil {
		p = &page.Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	p := &page.Page{Title: title, Body: []byte(body)}

	if err := p.Save(); err != nil {
		handleInternalServerError(w, err)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *page.Page) {
	tmplName := strings.Join([]string{tmpl, "html"}, ".")

	if err := templates.ExecuteTemplate(w, tmplName, p); err != nil {
		handleInternalServerError(w, err)
	}
}

func handleInternalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)

}

func getTitle(r *http.Request) (string, error) {
	viewAndPath := validPath.FindStringSubmatch(r.URL.Path)

	if viewAndPath == nil {
		path := validRestPath.FindStringSubmatch(r.URL.Path)
		return "", errors.New(fmt.Sprintf("Please provide a valid url: http://localhost:8080/%s/<<page>>", path[1]))
	}

	return viewAndPath[2], nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := getTitle(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fn(w, r, title)
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
