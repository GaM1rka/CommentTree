package service

import (
	"context"
	"testing"

	"commenttree/backend/internal/comment/domain"
	"commenttree/backend/internal/comment/repository"
)

func TestCommentServiceBuildsTreeAndSearchesInSubtree(t *testing.T) {
	ctx := context.Background()
	svc := NewCommentService(repository.NewMemoryCommentRepository())

	root, err := svc.Create(ctx, domain.CreateCommentInput{Author: "Alice", Text: "Root discussion"})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}

	reply, err := svc.Create(ctx, domain.CreateCommentInput{ParentID: &root.ID, Author: "Bob", Text: "Nested answer"})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}

	_, err = svc.Create(ctx, domain.CreateCommentInput{ParentID: &reply.ID, Author: "Carol", Text: "Needle is here"})
	if err != nil {
		t.Fatalf("create nested reply: %v", err)
	}

	page, err := svc.GetTree(ctx, domain.ListOptions{Query: "needle"})
	if err != nil {
		t.Fatalf("get tree: %v", err)
	}

	if page.Total != 1 {
		t.Fatalf("expected 1 matching root, got %d", page.Total)
	}
	if got := page.Items[0].Children[0].Children[0].Text; got != "Needle is here" {
		t.Fatalf("expected nested search result to keep subtree, got %q", got)
	}
}

func TestCommentServiceDeletesSubtree(t *testing.T) {
	ctx := context.Background()
	svc := NewCommentService(repository.NewMemoryCommentRepository())

	root, err := svc.Create(ctx, domain.CreateCommentInput{Author: "Alice", Text: "Root"})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	child, err := svc.Create(ctx, domain.CreateCommentInput{ParentID: &root.ID, Author: "Bob", Text: "Child"})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	_, err = svc.Create(ctx, domain.CreateCommentInput{ParentID: &child.ID, Author: "Carol", Text: "Grandchild"})
	if err != nil {
		t.Fatalf("create grandchild: %v", err)
	}

	deleted, err := svc.Delete(ctx, root.ID)
	if err != nil {
		t.Fatalf("delete subtree: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted comments, got %d", deleted)
	}

	page, err := svc.GetTree(ctx, domain.ListOptions{})
	if err != nil {
		t.Fatalf("get tree: %v", err)
	}
	if page.Total != 0 {
		t.Fatalf("expected empty tree, got %d roots", page.Total)
	}
}
