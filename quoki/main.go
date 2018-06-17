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
	"github.com/shyamz-22/gophercises/quoki/rand"
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
	viewTemplate = template.Must(template.New("templates/view.html").Funcs(template.FuncMap{"markDown": markDowner}).ParseFiles(
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

func viewHandler(w http.ResponseWriter, r *http.Request, id string) {
	p, err := page.LoadPage(id)
	if err != nil {
		http.Redirect(w, r, "/edit/"+rand.RandomString(32), http.StatusNotFound)
		return
	}

	d := struct {
		NewId string
		Page  *page.Page
	}{
		NewId: rand.RandomString(32),
		Page:  p,
	}
	if err := viewTemplate.ExecuteTemplate(w, "view.html", d); err != nil {
		handleInternalServerError(w, err)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, id string) {
	p, err := page.LoadPage(id)
	if err != nil {
		p = &page.Page{Id: id}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, id string) {
	body := r.FormValue("body")
	displayTitle := r.FormValue("displayTitle")

	p := &page.Page{Id: id, DisplayTitle: displayTitle, PagePath: page.GetAbsPath(id), Body: []byte(body)}

	if metaWriteError := p.WriteMetaData(); metaWriteError != nil {
		handleInternalServerError(w, metaWriteError)
		return
	}

	if err := p.Save(); err != nil {
		handleInternalServerError(w, err)
		return
	}

	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	tmplName := strings.Join([]string{"home", "html"}, ".")

	le := struct {
		NewId        string
		ListOfTitles page.Pages
	}{
		NewId:        rand.RandomString(32),
		ListOfTitles: page.ListPageTitles(),
	}

	if err := templates.ExecuteTemplate(w, tmplName, le); err != nil {
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
