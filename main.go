// A small and simple standalone ``pastebin'' web application.
// AGPL, 2014, Lauri Peltomäki
// http://www.gnu.org/licenses/agpl-3.0.html
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"pastecan/pbnf"
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

var (
	canPath  = os.Getenv("HOME") + "/.local/share/pastecan/"
	htmlPath = canPath + "htmls/"
	cssPath  = canPath + "styles/"

	pastePath string
	port      string

	templates = template.Must(template.ParseFiles(
		htmlPath+"paste.html",
		htmlPath+"view.html",
		htmlPath+"gopaste.html",
		htmlPath+"goview.html",
		htmlPath+"luapaste.html",
		htmlPath+"luaview.html",
	))
	validPath = regexp.MustCompile(`^/(view|goview|luaview)/([a-zA-Z]+)$`)
)

func init() {
	flag.StringVar(&pastePath, "d", "/tmp/pastecan/", "Pastes here.")
	flag.StringVar(&pastePath, "dir", "/tmp/pastecan/", "Pastes here.")
	flag.StringVar(&port, "p", "12022", "Port to use.")
	flag.StringVar(&port, "port", "12022", "Port to use.")

	flag.Parse()

	if _, err := os.Stat(pastePath); os.IsNotExist(err) {
		if err := os.Mkdir(pastePath, 0777); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create paste dir.")
			os.Exit(1)
		}
	}

	if !strings.HasSuffix(pastePath, "/") {
		pastePath += "/"
	}
}

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
		switch syn {
		case "go", "lua":
			body = pbnf.Colourify(syn, body)
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
	fmt.Printf("Started %s in port: %s\n", os.Args[0], port)
	http.HandleFunc(makeHandler("paste"))
	http.HandleFunc(makeHandler("gopaste"))
	http.HandleFunc(makeHandler("luapaste"))
	http.HandleFunc(makeSaveHandler("save", ""))
	http.HandleFunc(makeSaveHandler("savego", "go"))
	http.HandleFunc(makeSaveHandler("savelua", "lua"))
	http.HandleFunc(makeViewHandler("view"))
	http.HandleFunc(makeViewHandler("goview"))
	http.HandleFunc(makeViewHandler("luaview"))
	// For "external" CSS, you need to reveal a small
	// fraction of the filesystem, such horror !
	http.Handle(
		"/styles/",
		http.StripPrefix(
			"/styles/",
			http.FileServer(http.Dir(cssPath)),
		),
	)
	http.ListenAndServe(":"+port, nil)
}
