//go:build ignore

// sitegen - saytning statik nusxasini (gh-pages uchun) generatsiya qiladi.
// Ishga tushirish:  go run sitegen.go
package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Movie struct {
	Rank        int      `json:"rank"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Genre       []string `json:"genre"`
	Rating      float64  `json:"rating"`
	Duration    string   `json:"duration"`
	Director    string   `json:"director"`
	Description string   `json:"description"`
	Poster      string   `json:"poster"`
	Original    string   `json:"original"`
	Watch       string   `json:"watch"`
}

type PageData struct {
	Movies []Movie
	Genres []string
	DataJS template.JS
}

func main() {
	out := "gh-pages"
	os.MkdirAll(out, 0755)

	var movies []Movie
	raw, err := os.ReadFile("data/movies.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(raw, &movies); err != nil {
		log.Fatal(err)
	}
	for i := range movies {
		movies[i].Rank = i + 1
		movies[i].Watch = "https://uzbeklar.biz/?do=search&subaction=search&story=" + url.QueryEscape(movies[i].Title)
	}

	genreSet := map[string]bool{}
	for _, m := range movies {
		for _, g := range m.Genre {
			genreSet[g] = true
		}
	}
	genres := make([]string, 0, len(genreSet))
	for g := range genreSet {
		genres = append(genres, g)
	}
	for i := 1; i < len(genres); i++ {
		for j := i; j > 0 && genres[j] < genres[j-1]; j-- {
			genres[j], genres[j-1] = genres[j-1], genres[j]
		}
	}

	jsData, err := json.Marshal(movies)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("index.html").Funcs(template.FuncMap{
		"join": func(sep string, elems []string) string { return strings.Join(elems, sep) },
	}).ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	buf := new(strings.Builder)
	err = t.Execute(buf, PageData{Movies: movies, Genres: genres, DataJS: template.JS(jsData)})
	if err != nil {
		log.Fatal(err)
	}

	html := strings.ReplaceAll(buf.String(), "/static/", "static/")
	if err := os.WriteFile(filepath.Join(out, "index.html"), []byte(html), 0644); err != nil {
		log.Fatal(err)
	}

	if err := copyDir("static", filepath.Join(out, "static")); err != nil {
		log.Fatal(err)
	}

	os.WriteFile(filepath.Join(out, ".nojekyll"), []byte(""), 0644)

	log.Printf("Tayyor! %s papkasiga statik sayt yozildi.", out)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}