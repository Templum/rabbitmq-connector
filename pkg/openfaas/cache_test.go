package openfaas

import "testing"

func TestTopicMap(t *testing.T) {
	t.Parallel()

	t.Run("Should override cache with update", func(t *testing.T) {
		cache := NewTopicFunctionCache()
		update := make(map[string][]string)
		update["billing"] = []string{"taxes", "notify"}

		before := len(cache.topicMap)
		cache.Refresh(update)
		after := len(cache.topicMap)

		if before == after {
			t.Errorf("Expected cache content to be different after update")
		}
	})

	t.Run("Should return all found functions for topic", func(t *testing.T) {
		cache := NewTopicFunctionCache()
		update := make(map[string][]string)
		update["billing"] = []string{"taxes", "notify"}
		cache.Refresh(update)

		found := cache.GetCachedValues("billing")

		if len(found) != 2 {
			t.Errorf("Expected 2 entries for billing but received %d entries", len(found))
		}
	})

	t.Run("Should return empty list if topic does not exist", func(t *testing.T) {
		cache := NewTopicFunctionCache()

		found := cache.GetCachedValues("billing")

		if len(found) > 0 {
			t.Errorf("Expected empty list but received list with %d entries", len(found))
		}
	})
}
