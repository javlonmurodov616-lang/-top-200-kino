// Top 200 eng zo'r kinolar - Go veb-ilovasi
//
// Ushbu loyiha:
//   - Standart net/http paketi bilan web server
//   - html/template orqali HTML render qilish
//   - embed paketi orqali static fayllar (CSS, JS) va ma'lumotlarni o'rnatish
//   - 50 ta kino ma'lumoti data/movies.json faylida saqlanadi
//
// Ishga tushirish:  go run main.go  ->  http://localhost:8080

package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Ma'lumotlar tuzilmalari
// ---------------------------------------------------------------------------

// Movie - bitta kino haqidagi ma'lumotlar. JSON'da data/movies.json faylida
// saqlanadi va o'zgartirish juda oson.
type Movie struct {
	Rank        int      `json:"rank"`        // Kartada ko'rsatiladigan o'rni (1..50)
	Title       string   `json:"title"`       // Kino nomi (o'zbekcha)
	Year        int      `json:"year"`        // Chiqarilgan yili
	Genre       []string `json:"genre"`       // Janrlari (ro'yxat)
	Rating      float64  `json:"rating"`      // Reyting (IMDb, 10 ballik)
	Duration    string   `json:"duration"`    // Davomiyligi (masalan "2 soat 22 daqiqa")
	Director    string   `json:"director"`    // Rejissyori
	Description string   `json:"description"` // Qisqa tavsifi (karta va modalda)
	Poster      string   `json:"poster"`      // Poster URL (agar yuklanmasa, avtomatik fallback)
	Original    string   `json:"original"`    // Asl (inglizcha) nomi - IMDb qidiruvi uchun
	Watch       string   `json:"watch"`       // IMDb qidiruv linki (hisoblab chiqiladi)
}

// PageData - html/template'ga uzatiladigan barcha ma'lumotlar.
type PageData struct {
	Movies []Movie   // Server tomonidan render qilish uchun
	Genres []string  // Janr filtr tugmalari ro'yxati
	DataJS template.JS // JavaScript (modal/filtr/qidiruv) uchun JSON
}

// ---------------------------------------------------------------------------
// embed - static fayllar, template va ma'lumotlar dasturga o'rnatiladi
// ---------------------------------------------------------------------------

//go:embed static
var staticFiles embed.FS // CSS/JS fayllar

//go:embed templates
var templateFiles embed.FS // HTML template

//go:embed data/movies.json
var moviesJSON []byte // Kino ma'lumotlari

// tmpl - bitta bosh sahifa template'i (Fayl ishlatilganda bitta marta tahlil qilinadi)
var tmpl *template.Template

// main - dasturga kirish nuqtasi
func main() {
	// 1) JSON'ni o'qib, Movie ro'yxatiga aylantiramiz
	var movies []Movie
	if err := json.Unmarshal(moviesJSON, &movies); err != nil {
		log.Fatalf("data/movies.json ni o'qishda xato: %v", err)
	}

	// 2) O'rinni JSON tartibiga qarab beramiz (1 dan 200 gacha) va
	//    har kino uchun IMDb qidiruv linkini hisoblaymiz
	for i := range movies {
		movies[i].Rank = i + 1
		movies[i].Watch = "https://www.imdb.com/find/?q=" + url.QueryEscape(movies[i].Original) + "&s=tt&ttype=ft"
	}

	// 3) Janrlar ro'yxatini yig'amiz (filtrlash tugmalari uchun)
	genreSet := make(map[string]bool)
	for _, m := range movies {
		for _, g := range m.Genre {
			genreSet[g] = true
		}
	}
	genres := make([]string, 0, len(genreSet))
	for g := range genreSet {
		genres = append(genres, g)
	}
	sort.Strings(genres)

	// 4) JavaScript uchun JSON tayyorlaymiz (modal, filtr, qidiruv ishlatadi)
	jsData, err := json.Marshal(movies)
	if err != nil {
		log.Fatalf("JSON yaratishda xato: %v", err)
	}

	// 5) Template'ni tahlil qilamiz (embed ichidan)
	funcMap := template.FuncMap{
		"join": func(sep string, elems []string) string { return strings.Join(elems, sep) },
	}
	tmpl = template.Must(
		template.New("index.html").Funcs(funcMap).ParseFS(templateFiles, "templates/index.html"),
	)

	// 6) Static fayllar uchun fayl server
	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static papkani ochishda xato: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticSub))

	// -----------------------------------------------------------------------
	// HTTP marshrutlar
	// -----------------------------------------------------------------------

	// Bosh sahifa - HTML render qilinadi
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Faqat bosh sahifani (/) qabul qilamiz; qolganlari 404
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		data := PageData{
			Movies: movies,
			Genres: genres,
			DataJS: template.JS(jsData),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Template render xatosi: %v", err)
			http.Error(w, "Ichki server xatosi", http.StatusInternalServerError)
		}
	})

	// Static fayllar: /static/css/style.css, /static/js/main.js
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// 7) Portni aniqlaymiz (hosting PORT beradi, lokalda 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server ishga tushdi: http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Serverda xato: %v", err)
	}
}
