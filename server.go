package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func index(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s, to path: %s, from: %s", r.Method, r.URL.Path, r.RemoteAddr)
	http.ServeFile(w, r, filepath.Join("svelte", "src", "app.html"))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {

	http.HandleFunc("GET /", index)
	http.HandleFunc("GET /healthz", healthCheck)

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")
	logger.Println("Server is ready to handle requests at :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("ListenAndServe: ", err)
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on :8080: %v\n", err)
	}

	<-done
	logger.Println("Server stopped")
}
