package sdk

import (
	"testing"
	"sync"
	"github.com/openfaas/faas/gateway/requests"
)

func TestTopicFunctionMap_Match(t *testing.T) {
	topics := make(map[string][]string)
	topics["routing"] = []string{"agentChecker","customerTracker","customerHistory"}
	topics["billing"] = []string{"emailTranscript"}

	mapping := TopicFunctionMap{
		mutex: sync.Mutex{},
		lookupTable: &topics,
	}

	foundFunction := mapping.Match("routing")
	if len(foundFunction) != 3 {
		t.Errorf("Match result is wrong: Want %d received %d", 3, len(foundFunction))
	}

	foundFunction = mapping.Match("billing")
	if len(foundFunction) != 1 {
		t.Errorf("Match result is wrong: Want %d received %d", 1, len(foundFunction))
	}

	foundFunction = mapping.Match("not-existing")
	if len(foundFunction) > 0{
		t.Errorf("Match result is wrong: Want %d received %d", 0, len(foundFunction))
	}
}

func TestTopicFunctionMap_Sync(t *testing.T) {
	mapping := NewTopicFunctionMap()

	labelMap := make(map[string]string)
	labelMap["topic"] = "billing"

	sampleFunction := &[]requests.Function{
		{
			Name: "emailTranscript",
			Labels: &labelMap,
		},
	}

	foundFunction := mapping.Match("billing")
	if len(foundFunction) > 0{
		t.Errorf("Match result is wrong: Want %d received %d", 0, len(foundFunction))
	}

	mapping.Sync(sampleFunction)

	foundFunction = mapping.Match("billing")
	if len(foundFunction) != 1 {
		t.Errorf("Match result is wrong: Want %d received %d", 1, len(foundFunction))
	}

}