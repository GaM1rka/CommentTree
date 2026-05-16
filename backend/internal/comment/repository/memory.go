package repository

import (
	"context"
	"sync"
	"time"

	"commenttree/backend/internal/comment/domain"
)

type MemoryCommentRepository struct {
	mu       sync.RWMutex
	nextID   int64
	comments map[int64]domain.Comment
}

func NewMemoryCommentRepository() *MemoryCommentRepository {
	return &MemoryCommentRepository{
		nextID:   1,
		comments: make(map[int64]domain.Comment),
	}
}

func (r *MemoryCommentRepository) Create(_ context.Context, input domain.CreateCommentInput) (domain.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	comment := domain.Comment{
		ID:        r.nextID,
		ParentID:  input.ParentID,
		Author:    input.Author,
		Text:      input.Text,
		CreatedAt: time.Now().UTC(),
	}

	r.comments[comment.ID] = comment
	r.nextID++

	return comment, nil
}

func (r *MemoryCommentRepository) FindByID(_ context.Context, id int64) (domain.Comment, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, ok := r.comments[id]
	return comment, ok, nil
}

func (r *MemoryCommentRepository) List(_ context.Context) ([]domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comments := make([]domain.Comment, 0, len(r.comments))
	for _, comment := range r.comments {
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *MemoryCommentRepository) DeleteSubtree(_ context.Context, id int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := r.deleteSubtreeLocked(id)
	return removed, nil
}

func (r *MemoryCommentRepository) deleteSubtreeLocked(id int64) int {
	removed := 0

	for childID, comment := range r.comments {
		if comment.ParentID != nil && *comment.ParentID == id {
			removed += r.deleteSubtreeLocked(childID)
		}
	}

	if _, ok := r.comments[id]; ok {
		delete(r.comments, id)
		removed++
	}

	return removed
}
