package httpapi

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"

	"commenttree/backend/internal/comment/service"
)

//go:embed static/swagger.html
var swaggerHTML embed.FS

//go:embed static/openapi.yaml
var openAPIDoc embed.FS

type Router struct {
	comments *CommentHandler
	mux      *http.ServeMux
}

func NewRouter(commentService *service.CommentService) http.Handler {
	router := &Router{
		comments: NewCommentHandler(commentService),
		mux:      http.NewServeMux(),
	}

	router.routes()
	return withCORS(router.mux)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /health", r.health)
	r.mux.HandleFunc("GET /swagger", r.swagger)
	r.mux.HandleFunc("GET /openapi.yaml", r.openapi)
	r.mux.HandleFunc("POST /comments", r.comments.Create)
	r.mux.HandleFunc("GET /comments", r.comments.List)
	r.mux.HandleFunc("DELETE /comments/{id}", r.comments.Delete)
}

func (r *Router) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) swagger(w http.ResponseWriter, req *http.Request) {
	http.ServeFileFS(w, req, swaggerHTML, "static/swagger.html")
}

func (r *Router) openapi(w http.ResponseWriter, _ *http.Request) {
	spec, err := fs.ReadFile(openAPIDoc, "static/openapi.yaml")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "openapi spec is not available")
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, service.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
