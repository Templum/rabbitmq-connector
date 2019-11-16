package openfaas

import "testing"

func TestFunctionMapBuilder_Append(t *testing.T) {
	t.Parallel()

	t.Run("Should skip appending if topic is '' or ' '", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("", "NotIncluded")
		target.Append(" ", "NotIncluded")

		build := target.Build()

		if len(build) != 0 {
			t.Errorf("Expected map to be empty instead it contained %d elements", len(build))
		}
	})

	t.Run("Should append to existing entries", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("Billing", "CalcTax")
		target.Append("Billing", "NotifyLogistic")
		build := target.Build()

		if build["Billing"] == nil {
			t.Error("Expected map to contain entry for billing")
		}

		if len(build["Billing"]) != 2 {
			t.Errorf("Expected two entry for billing but received %d entries", len(build["Billing"]))
		}
	})
}

func TestFunctionMapBuilder_Build(t *testing.T) {
	t.Parallel()

	t.Run("Should return a empty map if nothing was appended", func(t *testing.T) {
		target := NewFunctionMapBuilder()
		build := target.Build()

		if len(build) != 0 {
			t.Errorf("Expected map to be empty instead it contained %d elements", len(build))
		}
	})

	t.Run("Should return a map based on previous append", func(t *testing.T) {
		target := NewFunctionMapBuilder()

		target.Append("Billing", "CalcTax")
		build := target.Build()

		if build["Billing"] == nil {
			t.Error("Expected map to contain entry for billing")
		}

		if len(build["Billing"]) != 1 {
			t.Errorf("Expected one entry for billing but received %d entries", len(build["Billing"]))
		}
	})
}
