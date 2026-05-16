package service

import (
	"sort"
	"strings"

	"commenttree/backend/internal/comment/domain"
)

func sortComments(comments []domain.Comment, field domain.SortField, order domain.SortOrder) {
	sort.SliceStable(comments, func(i, j int) bool {
		left := comments[i]
		right := comments[j]

		var cmp int
		switch field {
		case domain.SortAuthor:
			cmp = strings.Compare(strings.ToLower(left.Author), strings.ToLower(right.Author))
		case domain.SortText:
			cmp = strings.Compare(strings.ToLower(left.Text), strings.ToLower(right.Text))
		default:
			cmp = left.CreatedAt.Compare(right.CreatedAt)
		}

		if order == domain.OrderDesc {
			return cmp > 0
		}

		return cmp < 0
	})
}
