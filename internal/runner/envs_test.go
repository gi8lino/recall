package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvString(t *testing.T) {
	t.Run("returns environment value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_VALUE", "configured")

		assert.Equal(t, "configured", envString("RECALL_TEST_VALUE", "fallback"))
	})

	t.Run("returns fallback for empty value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_VALUE", "")

		assert.Equal(t, "fallback", envString("RECALL_TEST_VALUE", "fallback"))
	})
}

func TestEnvInt(t *testing.T) {
	t.Run("returns environment integer", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "42")

		assert.Equal(t, 42, envInt("RECALL_TEST_INT", 7))
	})

	t.Run("returns fallback for invalid integer", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "nope")

		assert.Equal(t, 7, envInt("RECALL_TEST_INT", 7))
	})

	t.Run("returns fallback for empty value", func(t *testing.T) {
		t.Setenv("RECALL_TEST_INT", "")

		assert.Equal(t, 7, envInt("RECALL_TEST_INT", 7))
	})
}
