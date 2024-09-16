package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	_, err := os.Stat("static")
	if os.IsNotExist(err) {
		unzip("static.zip")
	}

	Init()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/homepage/page_1", 302)
	})
	http.HandleFunc("/admin/", adminHandler)
	http.HandleFunc("/admin/login/", loginHandler)

	initArticles()

	fmt.Println("Starting server on port :8888...")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
