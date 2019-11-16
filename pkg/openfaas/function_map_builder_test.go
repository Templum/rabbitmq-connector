package openfaas

import (
	"gotest.tools/assert"
	"testing"
)

func TestFunctionMapBuilder_Append(t *testing.T) {
	t.Parallel()

	t.Run("Should skip appending if topic is '' or ' '", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("", "NotIncluded")
		target.Append(" ", "NotIncluded")

		build := target.Build()

		assert.Check(t, len(build) == 0, "Expected to skip empty/whitespace topic")
	})

	t.Run("Should append to existing entries", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("Billing", "CalcTax")
		target.Append("Billing", "NotifyLogistic")
		build := target.Build()

		assert.Check(t, build["Billing"] != nil, "Expected added Topic to be present")
		assert.Check(t, len(build["Billing"]) == 2, "Expected two entries")
	})
}

func TestFunctionMapBuilder_Build(t *testing.T) {
	t.Parallel()

	t.Run("Should return a empty map if nothing was appended", func(t *testing.T) {
		target := NewFunctionMapBuilder()
		build := target.Build()

		assert.Check(t, len(build) == 0, "Expected empty map when nothing was appended")
	})

	t.Run("Should return a map based on previous append", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("Billing", "CalcTax")
		build := target.Build()

		assert.Check(t, build["Billing"] != nil, "Expected added Topic to be present")
		assert.Check(t, len(build["Billing"]) == 1, "Expected one entry")
	})
}
