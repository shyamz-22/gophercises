package page

import (
	"io/ioutil"
	"path"
	"log"
	"os"
	"encoding/csv"
	"errors"
)

type Pages []*Page

type Page struct {
	Id           string
	DisplayTitle string
	PagePath     string
	Body         []byte
}

func readAll() Pages {
	base, _ := os.Getwd()

	metaFileLocation := path.Join(base, "meta", "meta.csv")
	f, err := os.Open(metaFileLocation)
	if err != nil {
		log.Printf("Unable to open file %s. Err %v", metaFileLocation, err)
		return nil
	}
	defer f.Close()

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	if err != nil {
		log.Printf("Invalid csv file %s. Err %v", metaFileLocation, err)
		return nil
	}

	titleToPageMappings := make([]*Page, len(lines))

	for i, tm := range lines {
		titleToPageMappings[i] = &Page{
			Id:           tm[0],
			PagePath:     tm[1],
			DisplayTitle: tm[2],
		}
	}

	return titleToPageMappings
}

func writeToMetaCsv(args ...string) error {
	var err error
	var f *os.File

	base, _ := os.Getwd()
	metaFileLocation := path.Join(base, "meta", "meta.csv")

	if f, err = os.OpenFile(metaFileLocation, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		log.Printf("Unable to open/create file %s. Err %v", metaFileLocation, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	err = writer.Write(args)
	if err != nil {
		log.Printf("Unable to open file %s. Err %v", metaFileLocation, err)
		return err
	}

	return nil
}

func (mappings Pages) getPageById(id string) *Page {
	for _, mapping := range mappings {

		if id == mapping.Id {
			return mapping
		}
	}

	return nil
}

func (p *Page) Save() error {

	var err error

	if err = ioutil.WriteFile(p.PagePath, p.Body, 0600); err != nil {
		return err
	}

	return nil
}

func (p *Page) WriteMetaData() error {

	if err := writeToMetaCsv(p.Id, p.PagePath, p.DisplayTitle); err != nil {
		return err
	}

	return nil
}

func LoadPage(id string) (*Page, error) {

	var (
		err   error
		tm    *Page
		body  []byte
		pages Pages
	)

	if pages = readAll(); pages == nil {
		return nil, errors.New("page not found: " + id)
	}

	if tm = pages.getPageById(id); tm == nil {
		return nil, errors.New("page not found: " + id)
	}

	if body, err = ioutil.ReadFile(tm.PagePath); err != nil {
		return nil, err
	}

	tm.Body = body

	return tm, nil
}

func ListPageTitles() Pages {
	return readAll()
}

func GetAbsPath(id string) string {
	base, _ := os.Getwd()
	filename := id + ".md"
	absolutePath := path.Join(base, "pages", filename)
	return absolutePath
}
