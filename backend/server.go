package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func dbConnection() (*mongo.Client, error) {
	dbLogger := log.New(os.Stdout, "db: ", log.LstdFlags)
	dbAddr := "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbAddr))
	if err != nil {
		dbLogger.Println("DB Connection error: ", err)
	}

	return client, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s, to path: %s, from: %s", r.Method, r.URL.Path, r.RemoteAddr)
	http.ServeFile(w, r, filepath.Join("public", "index.html"))
}

func archive(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s, to path: %s, from: %s", r.Method, r.URL.Path, r.RemoteAddr)
	http.ServeFile(w, r, filepath.Join("public", "archive", "archive.html"))
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	dbConnection()
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	type PageData struct {
		CurrentPage int
		PrevPage    int
		NextPage    int
	}

	pageParam := r.URL.Path[len("/archive/"):]
	currentPage, err := strconv.Atoi(pageParam)
	if err != nil || currentPage <= 0 {
		currentPage = 1 // Default to page 1 if the page number is invalid
	}

	// Calculate the previous and next page numbers
	prevPage := currentPage - 1
	logger.Printf("Current page: ", currentPage)
	if prevPage <= 0 {
		prevPage = 1 // Prevent going below page 1
	}
	nextPage := currentPage + 1

	// Example data , to be replaced by MongoDB posts.
	data := PageData{
		CurrentPage: currentPage,
		PrevPage:    prevPage,
		NextPage:    nextPage,
	}

	DBdata := PostData{
		PostTitle:   postTitle,
		PostContent: PostContent,
		PostID:      nextPage,
	}

	coll := DBdata.Database("db").Collection("students")

	address1 := DBdata{}

	_, err = coll.InsertOne(context.TODO()

	type PostD	ata struct {
		PostTitle   string `bson:"postTitle"`
		PostContent string `bson:"postContent"`
		PostID      int    `bson:"postID"`
	}

	// Parse the template
	tmpl, err := template.ParseFiles("public/archive/archivePosts.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s, to path: %s, from: %s", r.Method, r.URL.Path, r.RemoteAddr)
	http.ServeFile(w, r, filepath.Join("public", "login", "login.html"))
}

func aboutme(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s, to path: %s, from: %s", r.Method, r.URL.Path, r.RemoteAddr)
	http.ServeFile(w, r, filepath.Join("public", "aboutme", "aboutme.html"))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	wLogger := log.New(os.Stdout, "http: ", log.LstdFlags)

	//Connect to MongoDB
	client, err := dbConnection()

	//Error Checking
	if err != nil {
		wLogger.Println("DB Connection error: ", err)
	}
	defer client.Disconnect(context.Background()) //whenever webserer shuts down, so does the connection

	//authCollect := client.Database("GoDB").Collection("auth")

	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Address
	wAddr := "127.0.0.1:9000" //website addr

	//mux := http.NewServeMux()

	// Routes
	http.HandleFunc("/", index)
	http.HandleFunc("/archive", archive)
	http.HandleFunc("/login", login)
	http.HandleFunc("/aboutme", aboutme)
	http.HandleFunc("/healthz", healthCheck)

	http.HandleFunc("/archive/", postsHandler)

	//posts := client.Database("GoDB").Collection("posts")

	//Start Server
	wLogger.Println("Server is starting...")
	wLogger.Println("Server is ready to handle requests at:", wAddr)

	//Website error handling
	wErr := http.ListenAndServe(wAddr, nil)
	if wErr != nil {
		wLogger.Fatal("ListenAndServe: ", wErr)
	}

	// Graceful shutdown
	server := &http.Server{
		Addr:         wAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		wLogger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			wLogger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	wLogger.Println("Server is ready to handle requests at : ", wAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		wLogger.Fatalf("Could not listen on :8080: %v\n", err)
	}

	<-done
	wLogger.Println("Server stopped")
}
