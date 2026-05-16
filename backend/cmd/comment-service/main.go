package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"commenttree/backend/internal/comment/repository"
	"commenttree/backend/internal/comment/service"
	"commenttree/backend/internal/httpapi"
)

func main() {
	addr := getenv("HTTP_ADDR", ":8080")

	repo := repository.NewMemoryCommentRepository()
	commentService := service.NewCommentService(repo)
	router := httpapi.NewRouter(commentService)

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("comment-service is listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
