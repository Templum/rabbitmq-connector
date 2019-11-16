package openfaas

import (
	"gotest.tools/assert"
	"testing"
)

func TestTopicMap(t *testing.T) {
	t.Parallel()

	t.Run("Should override cache with update", func(t *testing.T) {
		cache := NewTopicFunctionCache()
		update := make(map[string][]string)
		update["billing"] = []string{"taxes", "notify"}

		before := len(cache.topicMap)
		cache.Refresh(update)
		after := len(cache.topicMap)

		assert.Check(t, before != after, "Expected that the update overrides the initial value")
	})

	t.Run("Should return all found functions for topic", func(t *testing.T) {
		cache := NewTopicFunctionCache()
		update := make(map[string][]string)
		update["billing"] = []string{"taxes", "notify"}
		cache.Refresh(update)

		found := cache.GetCachedValues("billing")
		assert.Check(t, len(found) == 2, "Expected 2 entries for billing")
	})

	t.Run("Should return empty list if topic does not exist", func(t *testing.T) {
		cache := NewTopicFunctionCache()

		found := cache.GetCachedValues("billing")
		assert.Check(t, len(found) == 0, "Expected empty list for non existing topic")
	})
}
