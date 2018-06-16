package page

import (
	"io/ioutil"
	"path"
	"log"
	"strings"
	"os"
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

func ListPageTitles() []string {
	base, _ := os.Getwd()
	files, err := ioutil.ReadDir(path.Join(base, "pages"))
	if err != nil {
		log.Println(err)
		return []string{}
	}

	var pageTitles = make([]string, len(files))
	for index, file := range files {
		if nameWithExtension := file.Name(); !strings.HasPrefix(nameWithExtension, ".") {
			pageTitles[index] = removeFileExtension(file)
		}
	}

	return pageTitles
}

func removeFileExtension(file os.FileInfo) string {
	fileName := strings.Split(file.Name(), ".")
	return fileName[0]
}

func getAbsPath(title string) string {
	filename := title + ".md"
	absolutePath := path.Join("pages", filename)
	return absolutePath
}
