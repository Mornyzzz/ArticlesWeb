package main

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Article struct {
	Title       string
	Description string
	Text        string
}

type AdminInfo struct {
	host       string
	port       int
	user       string
	password   string
	dbname     string
	adminLogin string
	adminPass  string
}

var access bool = false

var dbase *gorm.DB

func Init() *gorm.DB {
	info := getAdminData()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", info.host, info.port, info.user, info.password, info.dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if !db.Migrator().HasTable(&Article{}) {
		err = db.AutoMigrate(&Article{})
		if err != nil {
			log.Fatal(err)
		}
	}
	return db
}

func GetDB() *gorm.DB {
	if dbase == nil {
		dbase = Init()
	}
	return dbase
}

func homepageHandler(i int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := dataFromDatabase()
		for i := 0; i < len(data)%3; i++ {
			data = append(data, Article{
				Title:       "",
				Description: "",
				Text:        "",
			})
		}
		numPages := len(data) / 3
		htmlData := map[string]interface{}{
			"nowPage":       i,
			"lastPage":      numPages,
			"articleTitle1": data[i*3-3].Title,
			"articleDesc1":  data[i*3-3].Description,
			"articleTitle2": data[i*3-2].Title,
			"articleDesc2":  data[i*3-2].Description,
			"articleTitle3": data[i*3-1].Title,
			"articleDesc3":  data[i*3-1].Description,
		}
		pageTemplate, _ := template.ParseFiles("static/html/page.html")
		pageTemplate.Execute(w, htmlData)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	info := getAdminData()
	if r.Method == "GET" {
		// Отобразить страницу входа
		tmpl, _ := template.ParseFiles("static/html/login.html")
		tmpl.Execute(w, nil)
		return
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Проверка логина и пароля

		if username == info.adminLogin && password == info.adminPass {
			access = true
			http.Redirect(w, r, "/admin", 302)
			return
		} else {
			access = false
			tmpl, _ := template.ParseFiles("static/html/login.html")
			tmpl.Execute(w, nil)
		}
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	if access == true {
		if r.Method == "POST" {
			name := r.FormValue("name")
			desc := r.FormValue("desc")
			text := r.FormValue("text")
			if name == "" || desc == "" || text == "" {
				tmpl, _ := template.ParseFiles("static/html/adminPanel.html")
				tmpl.Execute(w, nil)
				return
			}
			dbase := GetDB()
			result := dbase.Create(Article{
				name,
				desc,
				text,
			})
			if result.Error != nil {
				log.Fatal(result.Error)
			}
			addArticle()
		}
		tmpl, _ := template.ParseFiles("static/html/adminPanel.html")
		tmpl.Execute(w, nil)
	} else {
		http.Redirect(w, r, "/admin/login/", 302)
	}
}

func articleHandler(page int, article int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := dataFromDatabase()
		for i := 0; i < len(data)%3; i++ {
			data = append(data, Article{
				Title:       "",
				Description: "",
				Text:        "",
			})
		}
		htmlData := map[string]interface{}{
			"nowPage":     page,
			"articleDesc": data[page*3-(4-article)].Description,
			"articleText": data[page*3-(4-article)].Text,
		}
		tmpl, _ := template.ParseFiles("static/html/article.html")
		tmpl.Execute(w, htmlData)
	}
}

func dataFromDatabase() []Article {
	GetDB()
	info := getAdminData()
	data := []Article{}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", info.host, info.port, info.user, info.password, info.dbname)
	// Создайте новый контекст с таймаутом 5 секунд.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Подключитесь к базе данных с помощью функции pgx.Connect().
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic("не удалось разобрать конфигурацию базы данных")
	}
	pgxConn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		panic("не удалось подключиться к базе данных")
	}
	// Закройте соединение, когда завершится функция main.
	defer pgxConn.Close(ctx)
	rows, err := pgxConn.Query(ctx, "SELECT title, description, text FROM articles")
	if err != nil {
		panic("не удалось выполнить SQL-запрос")
	}
	defer rows.Close()
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.Title, &article.Description, &article.Text); err != nil {
			panic("не удалось сканировать строку")
		}
		data = append(data, article)
	}
	return data
}

func initArticles() {
	data := dataFromDatabase()
	for i := 0; i < len(data)%3; i++ {
		data = append(data, Article{
			Title:       "",
			Description: "",
			Text:        "",
		})
	}
	numPages := len(data) / 3
	for i := 1; i <= numPages; i++ {
		http.HandleFunc(fmt.Sprintf("/homepage/page_%d/", i), homepageHandler(i))
		for j := 1; j <= 3; j++ {
			http.HandleFunc(fmt.Sprintf("/homepage/page_%d/article_%d/", i, j), articleHandler(i, j))
		}
	}
}

func addArticle() {
	data := dataFromDatabase()
	i := len(data)/3 + 1
	if len(data)%3 == 1 {
		http.HandleFunc(fmt.Sprintf("/homepage/page_%d/", i), homepageHandler(i))
		for j := 1; j <= 3; j++ {
			http.HandleFunc(fmt.Sprintf("/homepage/page_%d/article_%d/", i, j), articleHandler(i, j))
		}
	}
}

func getAdminData() AdminInfo {
	var info AdminInfo
	data, err := os.ReadFile("static/adminData/admin_credentials.txt")
	if err != nil {
		log.Fatal()
	}
	// Parse the data into environment variables
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		if err := os.Setenv(key, value); err != nil {
			log.Fatal()
		}
	}
	// Read database configuration from environment variables
	info.host = os.Getenv("HOST")
	info.port, err = strconv.Atoi(strings.TrimSuffix(os.Getenv("PORT"), "\r"))
	info.user = os.Getenv("USER")
	info.password = os.Getenv("PASSWORD")
	info.dbname = os.Getenv("DBNAME")

	// Read admin credentials from environment variables
	info.adminLogin = strings.TrimSpace(os.Getenv("ADMIN_PASS"))
	info.adminPass = strings.TrimSpace(os.Getenv("ADMIN_PASS"))

	return info
}

func unzip(filename string) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return
	}
	defer r.Close()

	dest := filepath.Dir(filename)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			dir := filepath.Dir(path)
			os.MkdirAll(dir, os.ModePerm)

			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return
			}
			defer file.Close()

			_, err = io.Copy(file, rc)
			if err != nil {
				return
			}
		}
	}
}
