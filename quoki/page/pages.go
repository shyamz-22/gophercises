package page

import (
	"io/ioutil"
	"path"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) Save() error {
	absolutePath := getAbsPath(p.Title)
	return ioutil.WriteFile(absolutePath, p.Body, 0600)
}

func LoadPage(title string) (*Page, error) {
	filename := getAbsPath(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getAbsPath(title string) string {
	filename := title + ".txt"
	absolutePath := path.Join("pages", filename)
	return absolutePath
}

