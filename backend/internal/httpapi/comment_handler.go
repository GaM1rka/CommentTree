package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"commenttree/backend/internal/comment/domain"
	"commenttree/backend/internal/comment/service"
)

type CommentHandler struct {
	service *service.CommentService
}

func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	comment, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	opts, err := parseListOptions(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	page, err := h.service.GetTree(r.Context(), opts)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	deleted, err := h.service.Delete(r.Context(), id)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"deleted": deleted})
}

func parseListOptions(r *http.Request) (domain.ListOptions, error) {
	query := r.URL.Query()
	opts := domain.ListOptions{
		Query: query.Get("q"),
		Sort:  domain.SortField(query.Get("sort")),
		Order: domain.SortOrder(query.Get("order")),
	}

	if parent := query.Get("parent"); parent != "" {
		parentID, err := strconv.ParseInt(parent, 10, 64)
		if err != nil || parentID < 1 {
			return domain.ListOptions{}, errInvalidQuery("parent must be a positive integer")
		}

		opts.ParentID = &parentID
	}

	page, err := parsePositiveInt(query.Get("page"), 1)
	if err != nil {
		return domain.ListOptions{}, errInvalidQuery("page must be a positive integer")
	}
	limit, err := parsePositiveInt(query.Get("limit"), 20)
	if err != nil {
		return domain.ListOptions{}, errInvalidQuery("limit must be a positive integer")
	}
	opts.Page = page
	opts.Limit = limit

	switch opts.Sort {
	case "", domain.SortCreatedAt, domain.SortAuthor, domain.SortText:
	default:
		return domain.ListOptions{}, errInvalidQuery("sort must be one of created_at, author, text")
	}
	switch opts.Order {
	case "", domain.OrderAsc, domain.OrderDesc:
	default:
		return domain.ListOptions{}, errInvalidQuery("order must be asc or desc")
	}

	return opts, nil
}

func parsePositiveInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, errInvalidQuery("")
	}

	return parsed, nil
}
