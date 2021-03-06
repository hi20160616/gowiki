package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	tmplPath = "tmpl"
	dataPath = "data"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filepath.Join(dataPath, filename), p.Body, 0600)
}

func (p *Page) Lines() []string {
	return strings.Split(string(p.Body), "\n")
}

func loadPage(title string) (*Page, error) {
	title, _ = url.QueryUnescape(title)
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filepath.Join(dataPath, filename))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	p.Body = interPageLink(p.Body)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	if err := p.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	viewHandler(w, r, "FrontPage")
}

// var templates = template.Must(template.ParseFiles(
//         filepath.Join(tmplPath, "lip_upper.html"),
//         filepath.Join(tmplPath, "lip_lower.html"),
//         filepath.Join(tmplPath, "navbar.html"),
//         filepath.Join(tmplPath, "view.html"),
//         filepath.Join(tmplPath, "edit.html")))
var templates = template.Must(template.ParseGlob(filepath.Join(tmplPath, "*.html")))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/(.+)$")

// pattern like `[!foobar]` means a inter-page need to be made as link
var interPage = regexp.MustCompile(`\[!.+\]`)

func interPageLink(body []byte) []byte {
	repl := func(pagename []byte) []byte {
		pagename = pagename[2 : len(pagename)-1]
		origin := pagename
		pagename = bytes.ReplaceAll(pagename, []byte(" "), []byte("-"))
		return []byte(fmt.Sprintf("<a href=\"/view/%s\">%s</a>", pagename, origin))
	}

	return interPage.ReplaceAllFunc(body, repl)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
