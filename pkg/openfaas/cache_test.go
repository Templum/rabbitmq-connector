/*
 * Copyright (c) Simon Pelczer 2021. All rights reserved.
 *  Licensed under the MIT license. See LICENSE file in the project root for full license information.
 */

package openfaas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicMap(t *testing.T) {
	t.Parallel()

	update := map[string][]string{"billing": {"taxes", "notify"}}

	t.Run("Should override cache with update", func(t *testing.T) {
		cache := NewTopicFunctionCache()

		before := len(cache.topicMap)
		cache.Refresh(update)
		after := len(cache.topicMap)

		assert.NotEqual(t, after, before, "Expected that the update overrides the initial value")
	})

	t.Run("Should return all found functions for topic", func(t *testing.T) {
		cache := NewTopicFunctionCache()
		cache.Refresh(update)

		found := cache.GetCachedValues("billing")
		assert.Len(t, found, 2, "Expected 2 entries for billing")
	})

	t.Run("Should return empty list if topic does not exist", func(t *testing.T) {
		cache := NewTopicFunctionCache()

		found := cache.GetCachedValues("billing")
		assert.Len(t, found, 0, "Expected empty list for non existing topic")
	})
}
