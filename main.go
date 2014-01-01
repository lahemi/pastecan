// A small and simple standalone ``pastebin'' web application.
// AGPL, 2014, Lauri Peltomäki
// http://www.gnu.org/licenses/agpl-3.0.html
package main

import (
	"pastecan/synsgo"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// We use template.HTML for body, so that we may still use templates
// easily, but without all those nasty automatic escaping.
type Page struct {
	Title string
	Body  template.HTML
}

var canPath = os.Getenv("HOME") + "/.local/share/pastecan/"
var htmlPath = canPath + "htmls/"
var pastePath = "/tmp/pastecan/"

var templates = template.Must(template.ParseFiles(
	htmlPath+"paste.html",
	htmlPath+"view.html",
	htmlPath+"gopaste.html",
	htmlPath+"goview.html",
))
var validPath = regexp.MustCompile(`^/(view|goview)/([a-zA-Z]+)$`)

// Hastebin-style: uneven -> consonant, even -> vowel.
func genRandTitle() (title string) {
	co := "bcdfghjklmnpqrstvwxyz"
	vo := "aeiou"
	var l []string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var st string = ""
	for i := 0; i < 12; i++ {
		if (i % 2) == 0 {
			if st != "" {
				l = append(l, st)
			}
			st = string(co[r.Intn(len(co))])
		} else {
			st = st + string(vo[r.Intn(len(vo))])
		}

	}
	for _, v := range l {
		title += v
	}
	return
}

func (p *Page) save() (*Page, error) {
	filename := pastePath + p.Title
	/*
		This does not work properly. Sometimes it
		catches an existing file once, and no more
		and sometimes not at all.
		On option could be to keep a list of reserved
		names inside the program or some such...
	*/
	_, e := os.Stat(filename)
	if e == nil {
		filename = pastePath + genRandTitle()
		for {
			_, err := os.Stat(filename)
			if err != nil {
				break
			}
			filename = pastePath + genRandTitle()
		}
		p.Title = strings.Split(filename, "/")[0]
	}
	err := ioutil.WriteFile(filename, []byte(p.Body), 0600)
	if err != nil {
		return p, err
	}
	return p, nil
}

func loadPage(title string) (*Page, error) {
	filename := pastePath + title
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func makeHandler(base string) (string, http.HandlerFunc) {
	path := "/" + base + "/"
	return path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		err := templates.ExecuteTemplate(w, base+".html", &Page{Title: base})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func makeViewHandler(base string) (string, http.HandlerFunc) {
	path := "/" + base + "/"
	return path, func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		p, err := loadPage(m[2])
		if err != nil {
			return
		}
		err = templates.ExecuteTemplate(w, base+".html", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func makeSaveHandler(base, syn string) (string, http.HandlerFunc) {
	path := "/" + base + "/"
	return path, func(w http.ResponseWriter, r *http.Request) {
		body := r.FormValue("body")
		title := genRandTitle()
		switch {
		case syn == "go":
			body = synsgo.Colourify(body)
		}
		p := &Page{Title: title, Body: template.HTML(body)}
		p, err := p.save()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/"+syn+"view/"+p.Title, http.StatusFound)
	}
}

func main() {
	http.HandleFunc(makeHandler("paste"))
	http.HandleFunc(makeHandler("gopaste"))
	http.HandleFunc(makeSaveHandler("save", ""))
	http.HandleFunc(makeSaveHandler("savego", "go"))
	http.HandleFunc(makeViewHandler("view"))
	http.HandleFunc(makeViewHandler("goview"))
	http.ListenAndServe(":12022", nil)
}
