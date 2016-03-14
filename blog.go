package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type Post struct {
	Title            string
	HyperLinkedTitle string
	Body             []byte
	Date             string
}

type Page struct {
	Posts  []*Post
	Limit  int
	Offset int
}

var templates = template.Must(template.ParseFiles("view.html"))

func main() {
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.ListenAndServe(":8080", Log(http.DefaultServeMux))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func loadPost(path string) (*Post, error) {
	filename := path
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Post{Title: path, Body: body}, nil
}

func getTitle(path string) (title, date string) {
	pathArray := strings.Split(path, "/")
	dateTitleArray := strings.Split(pathArray[len(pathArray)-1], ".")
	titleArray := strings.Split(dateTitleArray[0], "__")
	if len(titleArray) == 2 {
		title = strings.Replace(titleArray[1], "_", " ", -1)
	} else {
		title = strings.Replace(titleArray[0], "_", " ", -1)
	}
	date = strings.Replace(titleArray[0], "_", "-", -1)
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):]
	p, err := loadPost(path)
	if err != nil {
		fmt.Println("Error loading page "+path, err.Error())
		http.Redirect(w, r, "/view/Not_Found.html", http.StatusFound)
		return
	}
	title, date := getTitle(path)
	p.Title = title
	pathArray := strings.Split(path, "/")
	if len(pathArray) == 2 {
		p.HyperLinkedTitle = `<a href="` + strings.Replace(pathArray[1], " ", "_", -1) + `">` + title + `</a>`
	} else {
		p.HyperLinkedTitle = `<a href="` + strings.Replace(pathArray[0], " ", "_", -1) + `">` + title + `</a>`
	}
	p.Body = []byte(strings.Replace(string(p.Body), "\n", "<br/>", -1))
	p.Date = date
	renderTemplate(w, "view", p)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./posts")
	if err != nil {
		fmt.Println("Error getting list of file names", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileNames := make([]string, 0, 10000)
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	var posts []*Post
	for i := len(fileNames) - 1; i >= 0; i-- {
		p, err := loadPost("posts/" + fileNames[i])
		p.Body = []byte(strings.Replace(string(p.Body), "\n", "<br/>", -1))
		if err != nil {
			fmt.Println("Error getting list of posts", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bodySplit := strings.Split(string(p.Body), " ")
		if len(bodySplit) > 150 {
			p.Body = []byte(strings.Join(bodySplit[:150], " ") + "... ")
		}
		p.Body = append(p.Body, []byte(`<br/><br/><a href=/view/`+p.Title+`>Read more</a>`)...)
		title, date := getTitle(p.Title)
		p.Title = `<a href=/view/` + p.Title + `>` + title + `</a>`
		p.Date = date
		posts = append(posts, p)
	}
	t, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("Error parsing template", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, posts)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Post) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
