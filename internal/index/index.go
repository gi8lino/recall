package index

import (
	"strings"

	"github.com/containeroo/recall/internal/command"
)

// Index contains commands prepared for searching.
type Index struct {
	Items []command.Command // Items contains deduplicated commands in priority order.
}

// NewIndex creates a searchable command index.
func NewIndex(items []command.Command) Index {
	return Index{
		Items: dedupe(items),
	}
}

// Search returns commands matching the query while preserving priority order.
func (i Index) Search(query string, limit int) []command.Command {
	if limit < 0 {
		limit = 20
	}

	results := make([]command.Command, 0, len(i.Items))
	for _, item := range i.Items {
		if item.IsZero() || !item.Match(query) {
			continue
		}

		results = append(results, item)
		if limit > 0 && len(results) >= limit {
			break
		}
	}

	return results
}

// dedupe removes duplicate commands while preserving the first occurrence.
func dedupe(items []command.Command) []command.Command {
	seen := make(map[string]struct{}, len(items))
	result := make([]command.Command, 0, len(items))

	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Text))
		if key == "" {
			continue
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, item)
	}

	return result
}
