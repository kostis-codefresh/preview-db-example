package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	_ "github.com/lib/pq"
)

type ExampleApplication struct {
	AppVersion     string
	BackendVersion string
	DatabaseHost   string
	DatabasePort   string
	DB             *sql.DB
}

func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	exampleApp := ExampleApplication{}

	exampleApp.AppVersion = os.Getenv("APP_VERSION")
	if len(exampleApp.AppVersion) == 0 {
		exampleApp.AppVersion = "dev"
	}

	exampleApp.DatabaseHost = os.Getenv("DATABASE_HOST")
	if len(exampleApp.DatabaseHost) == 0 {
		exampleApp.DatabaseHost = "localhost"
	}

	exampleApp.DatabasePort = os.Getenv("DATABASE_PORT")
	if len(exampleApp.DatabasePort) == 0 {
		exampleApp.DatabasePort = "5432"
	}

	// Kubernetes check if app is ok
	http.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "up")
	})

	// Kubernetes check if app can serve requests
	http.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "yes")
	})

	http.HandleFunc("/", exampleApp.serveFiles)

	fmt.Printf("Demo application version %s is listening now at port %s\n", exampleApp.AppVersion, port)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}

func (exampleApp *ExampleApplication) serveFiles(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	p := "." + upath
	if p == "./" {
		exampleApp.home(w, r)
		return
	} else {
		p = filepath.Join("./static/", path.Clean(upath))
	}
	http.ServeFile(w, r, p)
}

func (exampleApp *ExampleApplication) connectToDB() {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", "postgres", "", "")

	var err error
	fmt.Print("Connecting to DB...")
	exampleApp.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	err = exampleApp.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OK")
}

func (exampleApp *ExampleApplication) home(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("./static/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	err = t.Execute(w, exampleApp)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}
}
