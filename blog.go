package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

type Post struct {
	Title string
	Body  []byte
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
	http.ListenAndServe(":8080", nil)
}

func loadPost(path string) (*Post, error) {
	filename := path
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Post{Title: path, Body: body}, nil
}

func getTitle(path string) string {
	pathArray := strings.Split(path, "/")
	titleArray := strings.Split(pathArray[len(pathArray)-1], ".")
	title := strings.Replace(titleArray[0], "_", " ", -1)
	return title
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):]
	p, err := loadPost(path)
	if err != nil {
		fmt.Println("Error loading page "+path, err.Error())
		http.Redirect(w, r, "/view/Not Found.html", http.StatusFound)
		return
	}
	fmt.Println(path, p.Title)
	p.Title = `<a href="` + strings.Replace(getTitle(path), " ", "_", -1) + `">` + getTitle(p.Title) + `</a>`
	renderTemplate(w, "view", p)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./posts")
	if err != nil {
		fmt.Println("Error getting list of file names", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var posts []*Post
	for _, file := range files {
		p, err := loadPost("posts/" + file.Name())
		if err != nil {
			fmt.Println("Error getting list of posts", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bodySplit := strings.Split(string(p.Body), " ")
		if len(bodySplit) > 150 {
			p.Body = []byte(strings.Join(bodySplit[:150], " ") + "... ")
		}
		p.Body = append(p.Body, []byte(`<a href=/view/`+p.Title+`>Read more</a>`)...)
		p.Title = getTitle(p.Title)
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
