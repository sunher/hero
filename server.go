// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
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
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request, title string) {
	loginTemplate(w, title)
}
func messageHandler(w http.ResponseWriter, r *http.Request, title string) {
	r.ParseForm()
	fmt.Println(r.Method)
	if r.Method == "GET" {
		fmt.Fprintf(w, "This is a GET request")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Println("Recived info:", r.Form)
		tracefile(r.Form.Get("info"))
		body, err := ioutil.ReadFile("message.txt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(body))
	}
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "login.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, "login.html", tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|login|message)/([a-zA-Z0-9]+)$")

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

//打印内容到文件中
//tracefile(fmt.Sprintf("receive:%s",v))
func tracefile(str_content string) {
	fd, _ := os.OpenFile("message.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	//fd_time:=time.Now().Format("2006-01-02 15:04:05");
	fd_content := strings.Join([]string{str_content, "\n"}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/login/", makeHandler(loginHandler))
	http.HandleFunc("/message/", makeHandler(messageHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
