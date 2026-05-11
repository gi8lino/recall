package index

import (
	"testing"

	"github.com/containeroo/recall/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	t.Run("deduplicates commands while preserving first occurrence", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			command.NewCommand("kubectl get pods", "pin", true),
			command.NewCommand("go test ./...", "history", false),
			command.NewCommand("KUBECTL GET PODS", "history", false),
		}

		idx := NewIndex(items)

		require.Len(t, idx.Items, 2)
		assert.Equal(t, "kubectl get pods", idx.Items[0].Text)
		assert.True(t, idx.Items[0].Pinned)
		assert.Equal(t, "go test ./...", idx.Items[1].Text)
	})

	t.Run("drops empty commands", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			command.NewCommand("", "history", false),
			command.NewCommand("   ", "history", false),
			command.NewCommand("go test ./...", "history", false),
		}

		idx := NewIndex(items)

		require.Len(t, idx.Items, 1)
		assert.Equal(t, "go test ./...", idx.Items[0].Text)
	})
}

func TestIndexSearch(t *testing.T) {
	t.Parallel()

	t.Run("returns all commands when query is empty", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommand("kubectl get pods", "pin", true),
			command.NewCommand("go test ./...", "history", false),
		})

		results := idx.Search("", 0)

		require.Len(t, results, 2)
		assert.Equal(t, "kubectl get pods", results[0].Text)
		assert.Equal(t, "go test ./...", results[1].Text)
	})

	t.Run("filters commands by query while preserving order", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommand("kubectl get pods", "pin", true),
			command.NewCommand("docker compose up", "history", false),
			command.NewCommand("kubectl describe pod nginx", "history", false),
		})

		results := idx.Search("kubectl", 0)

		require.Len(t, results, 2)
		assert.Equal(t, "kubectl get pods", results[0].Text)
		assert.Equal(t, "kubectl describe pod nginx", results[1].Text)
	})

	t.Run("matches descriptions", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommandWithDescription("kubectl get pods", "Show running pods", "pin", true),
			command.NewCommand("docker compose up", "history", false),
		})

		results := idx.Search("running", 0)

		require.Len(t, results, 1)
		assert.Equal(t, "kubectl get pods", results[0].Text)
	})

	t.Run("requires all query terms to match", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommandWithDescription("kubectl get pods", "Show running pods", "pin", true),
			command.NewCommand("kubectl get services", "history", false),
			command.NewCommand("docker compose up", "history", false),
		})

		results := idx.Search("kubectl pods", 0)

		require.Len(t, results, 1)
		assert.Equal(t, "kubectl get pods", results[0].Text)
	})

	t.Run("applies positive limit", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommand("echo one", "history", false),
			command.NewCommand("echo two", "history", false),
			command.NewCommand("echo three", "history", false),
		})

		results := idx.Search("echo", 2)

		require.Len(t, results, 2)
		assert.Equal(t, "echo one", results[0].Text)
		assert.Equal(t, "echo two", results[1].Text)
	})

	t.Run("treats zero limit as unlimited", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex([]command.Command{
			command.NewCommand("echo one", "history", false),
			command.NewCommand("echo two", "history", false),
			command.NewCommand("echo three", "history", false),
		})

		results := idx.Search("echo", 0)

		require.Len(t, results, 3)
		assert.Equal(t, "echo one", results[0].Text)
		assert.Equal(t, "echo two", results[1].Text)
		assert.Equal(t, "echo three", results[2].Text)
	})

	t.Run("treats negative limit as default limit", func(t *testing.T) {
		t.Parallel()

		items := make([]command.Command, 0, 25)
		for i := 0; i < 25; i++ {
			items = append(items, command.NewCommand("echo item", "history", false))
		}

		for i := range items {
			items[i].Text = items[i].Text + string(rune('a'+i))
		}

		idx := NewIndex(items)
		results := idx.Search("echo", -1)

		require.Len(t, results, 20)
	})

	t.Run("skips zero commands", func(t *testing.T) {
		t.Parallel()

		idx := Index{
			Items: []command.Command{
				command.NewCommand("", "history", false),
				command.NewCommand("go test ./...", "history", false),
			},
		}

		results := idx.Search("", 0)

		require.Len(t, results, 1)
		assert.Equal(t, "go test ./...", results[0].Text)
	})
}

func TestDedupe(t *testing.T) {
	t.Parallel()

	t.Run("preserves first duplicate", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			command.NewCommandWithDescription("kubectl get pods", "Pinned pods", "pin", true),
			command.NewCommandWithDescription("kubectl get pods", "History pods", "history", false),
		}

		results := dedupe(items)

		require.Len(t, results, 1)
		assert.Equal(t, "kubectl get pods", results[0].Text)
		assert.Equal(t, "Pinned pods", results[0].Description)
		assert.True(t, results[0].Pinned)
	})

	t.Run("deduplicates case insensitively", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			command.NewCommand("KUBECTL GET PODS", "pin", true),
			command.NewCommand("kubectl get pods", "history", false),
		}

		results := dedupe(items)

		require.Len(t, results, 1)
		assert.Equal(t, "KUBECTL GET PODS", results[0].Text)
	})

	t.Run("deduplicates after trimming whitespace", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			{Text: "  kubectl get pods  ", Source: "pin", Pinned: true},
			{Text: "kubectl get pods", Source: "history", Pinned: false},
		}

		results := dedupe(items)

		require.Len(t, results, 1)
		assert.Equal(t, "  kubectl get pods  ", results[0].Text)
	})

	t.Run("removes empty commands", func(t *testing.T) {
		t.Parallel()

		items := []command.Command{
			command.NewCommand("", "history", false),
			command.NewCommand("   ", "history", false),
			command.NewCommand("go test ./...", "history", false),
		}

		results := dedupe(items)

		require.Len(t, results, 1)
		assert.Equal(t, "go test ./...", results[0].Text)
	})
}
