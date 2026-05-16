package domain

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parentId,omitempty"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommentNode struct {
	Comment
	Children []CommentNode `json:"children"`
}

type CreateCommentInput struct {
	ParentID *int64 `json:"parentId"`
	Author   string `json:"author"`
	Text     string `json:"text"`
}

type ListOptions struct {
	ParentID *int64
	Query    string
	Sort     SortField
	Order    SortOrder
	Page     int
	Limit    int
}

type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type SortField string

const (
	SortCreatedAt SortField = "created_at"
	SortAuthor    SortField = "author"
	SortText      SortField = "text"
)

type SortOrder string

const (
	OrderAsc  SortOrder = "asc"
	OrderDesc SortOrder = "desc"
)
