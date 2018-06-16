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
	"os"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
)

const authReqMessage = "Authentication required"

var (
	templates = template.Must(
		template.ParseFiles(
			"templates/header.html",
			"templates/navbar.html",
			"templates/footer.html",
			"templates/edit.html",
			"templates/home.html"))
	viewTemplate  = template.Must(template.New("templates/view.html").Funcs(template.FuncMap{"markDown": markDowner}).ParseFiles(
		"templates/header.html",
		"templates/navbar.html",
		"templates/footer.html",
		"templates/view.html"))
	validRestPath = regexp.MustCompile("^/(edit|save|view)/.*")
	validPath     = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

func markDowner(body []byte) template.HTML {

	unsafeHtml := blackfriday.Run(body)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafeHtml)

	return template.HTML(html)

}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := page.LoadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusNotFound)
		return
	}

	if err := viewTemplate.ExecuteTemplate(w, "view.html", p); err != nil {
		handleInternalServerError(w, err)
	}
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

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	tmplName := strings.Join([]string{"home", "html"}, ".")

	if err := templates.ExecuteTemplate(w, tmplName, page.ListPageTitles()); err != nil {
		log.Println(err)
		handleInternalServerError(w, err)
		return
	}
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

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if username, password, ok := r.BasicAuth(); ok {
			if validUser(username, password) {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", "Basic realm=\"user\"")
		http.Error(w, authReqMessage, http.StatusUnauthorized)
	}
}

func validUser(username string, password string) bool {
	return username == os.Getenv("QUOKI_USERNAME") && password == os.Getenv("QUOKI_PASSWORD")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/view/", authenticate(makeHandler(viewHandler)))
	mux.HandleFunc("/edit/", authenticate(makeHandler(editHandler)))
	mux.HandleFunc("/save/", authenticate(makeHandler(saveHandler)))
	mux.HandleFunc("/", homePageHandler)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
