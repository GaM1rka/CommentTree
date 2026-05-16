package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"commenttree/backend/internal/comment/domain"
)

type CommentRepository interface {
	Create(ctx context.Context, input domain.CreateCommentInput) (domain.Comment, error)
	FindByID(ctx context.Context, id int64) (domain.Comment, bool, error)
	List(ctx context.Context) ([]domain.Comment, error)
	DeleteSubtree(ctx context.Context, id int64) (int, error)
}

type CommentService struct {
	repo CommentRepository
}

func NewCommentService(repo CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) Create(ctx context.Context, input domain.CreateCommentInput) (domain.Comment, error) {
	input.Author = strings.TrimSpace(input.Author)
	input.Text = strings.TrimSpace(input.Text)

	if input.Author == "" {
		return domain.Comment{}, fmt.Errorf("%w: author is required", ErrValidation)
	}
	if input.Text == "" {
		return domain.Comment{}, fmt.Errorf("%w: text is required", ErrValidation)
	}
	if input.ParentID != nil {
		if _, ok, err := s.repo.FindByID(ctx, *input.ParentID); err != nil {
			return domain.Comment{}, err
		} else if !ok {
			return domain.Comment{}, ErrNotFound
		}
	}

	return s.repo.Create(ctx, input)
}

func (s *CommentService) GetTree(ctx context.Context, opts domain.ListOptions) (domain.Page[domain.CommentNode], error) {
	opts = normalizeListOptions(opts)

	comments, err := s.repo.List(ctx)
	if err != nil {
		return domain.Page[domain.CommentNode]{}, err
	}

	roots := filterRoots(comments, opts.ParentID)
	if opts.Query != "" {
		roots = filterSearchMatches(comments, roots, opts.Query)
	}

	sortComments(roots, opts.Sort, opts.Order)
	tree := buildForest(comments, roots, opts)

	total := len(tree)
	start, end := pageBounds(total, opts.Page, opts.Limit)
	if start > total {
		tree = []domain.CommentNode{}
	} else {
		tree = tree[start:end]
	}

	return domain.Page[domain.CommentNode]{
		Items:      tree,
		Page:       opts.Page,
		Limit:      opts.Limit,
		Total:      total,
		TotalPages: totalPages(total, opts.Limit),
	}, nil
}

func (s *CommentService) Delete(ctx context.Context, id int64) (int, error) {
	if _, ok, err := s.repo.FindByID(ctx, id); err != nil {
		return 0, err
	} else if !ok {
		return 0, ErrNotFound
	}

	return s.repo.DeleteSubtree(ctx, id)
}

func normalizeListOptions(opts domain.ListOptions) domain.ListOptions {
	if opts.Sort == "" {
		opts.Sort = domain.SortCreatedAt
	}
	if opts.Order == "" {
		opts.Order = domain.OrderDesc
	}
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit < 1 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	opts.Query = strings.TrimSpace(opts.Query)
	return opts
}

func filterRoots(comments []domain.Comment, parentID *int64) []domain.Comment {
	roots := make([]domain.Comment, 0)

	for _, comment := range comments {
		if sameParent(comment.ParentID, parentID) {
			roots = append(roots, comment)
		}
	}

	return roots
}

func filterSearchMatches(comments []domain.Comment, roots []domain.Comment, query string) []domain.Comment {
	query = strings.ToLower(query)
	descendantsByParent := groupByParent(comments)
	matchedRoots := make([]domain.Comment, 0, len(roots))

	for _, root := range roots {
		if subtreeMatches(root, descendantsByParent, query) {
			matchedRoots = append(matchedRoots, root)
		}
	}

	return matchedRoots
}

func subtreeMatches(root domain.Comment, descendantsByParent map[int64][]domain.Comment, query string) bool {
	if commentMatches(root, query) {
		return true
	}

	for _, child := range descendantsByParent[root.ID] {
		if subtreeMatches(child, descendantsByParent, query) {
			return true
		}
	}

	return false
}

func buildForest(comments []domain.Comment, roots []domain.Comment, opts domain.ListOptions) []domain.CommentNode {
	descendantsByParent := groupByParent(comments)
	forest := make([]domain.CommentNode, 0, len(roots))

	for _, root := range roots {
		forest = append(forest, buildNode(root, descendantsByParent, opts))
	}

	return forest
}

func buildNode(comment domain.Comment, descendantsByParent map[int64][]domain.Comment, opts domain.ListOptions) domain.CommentNode {
	children := descendantsByParent[comment.ID]
	sortComments(children, opts.Sort, opts.Order)

	node := domain.CommentNode{
		Comment:  comment,
		Children: make([]domain.CommentNode, 0, len(children)),
	}

	for _, child := range children {
		node.Children = append(node.Children, buildNode(child, descendantsByParent, opts))
	}

	return node
}

func groupByParent(comments []domain.Comment) map[int64][]domain.Comment {
	groups := make(map[int64][]domain.Comment)

	for _, comment := range comments {
		if comment.ParentID == nil {
			continue
		}

		groups[*comment.ParentID] = append(groups[*comment.ParentID], comment)
	}

	return groups
}

func sameParent(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return *left == *right
}

func commentMatches(comment domain.Comment, query string) bool {
	text := strings.ToLower(comment.Text)
	author := strings.ToLower(comment.Author)
	return strings.Contains(text, query) || strings.Contains(author, query)
}

func pageBounds(total, page, limit int) (int, int) {
	start := (page - 1) * limit
	if start > total {
		return total + 1, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return start, end
}

func totalPages(total, limit int) int {
	if total == 0 {
		return 0
	}

	return int(math.Ceil(float64(total) / float64(limit)))
}
