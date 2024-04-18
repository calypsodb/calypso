package main

import (
	"html/template"
	"net/http"
)

type PageData struct {
	Title string
	Name  string
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Welcome",
		Name:  "John Doe",
	}

	tmpl, err := template.New("index").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>Hello, {{.Name}}!</h1>
	</body>
	</html>
	`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}