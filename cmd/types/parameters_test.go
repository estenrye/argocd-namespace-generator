package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var ns = NamespaceInfo{
	Labels: map[string]string{
		"key1": "value1",
		"key2": "value2",
	},
	Annotations: map[string]string{
		"annotation1": "value1",
		"annotation2": "value2",
	},
}

var value2 string = "value2"
var notpresent string = "notPresent"

func TestParametersMatchesLabels(t *testing.T) {
	params := Parameters{
		MatchLabels: []HasKeyValue{
			{Key: "key1", Value: nil},
			{Key: "key2", Value: &value2},
		},
	}

	assert.True(t, params.Matches(ns), "Parameters should match the namespace labels")

	params = Parameters{
		MatchLabels: []HasKeyValue{
			{Key: "key1", Value: &value2},
		},
	}
	assert.False(t, params.Matches(ns), "Parameters should not match the namespace labels due to value mismatch")
}

func TestParametersMatchesAnnotations(t *testing.T) {
	params := Parameters{
		MatchAnnotations: []HasKeyValue{
			{Key: "annotation1", Value: nil},
			{Key: "annotation2", Value: &value2},
		},
	}

	assert.True(t, params.Matches(ns), "Parameters should match the namespace annotations")

	params = Parameters{
		MatchAnnotations: []HasKeyValue{
			{Key: "annotation1", Value: &value2},
		},
	}
	assert.False(t, params.Matches(ns), "Parameters should not match the namespace annotations due to value mismatch")
}

func TestParametersExcludesLabels(t *testing.T) {
	params := Parameters{
		ExcludeLabels: []HasKeyValue{
			{Key: "key1", Value: nil},
			{Key: "key2", Value: &notpresent},
		},
	}
	assert.False(t, params.Matches(ns), "Parameters should not match the namespace due to excluded label")

	params = Parameters{
		ExcludeLabels: []HasKeyValue{
			{Key: "key1", Value: &value2},
		},
	}
	assert.True(t, params.Matches(ns), "Parameters should match the namespace due to excluded label")
}

func TestParametersExcludesAnnotations(t *testing.T) {
	params := Parameters{
		ExcludeAnnotations: []HasKeyValue{
			{Key: "annotation1", Value: nil},
			{Key: "annotation2", Value: &notpresent},
		},
	}
	assert.False(t, params.Matches(ns), "Parameters should not match the namespace due to excluded annotation")

	params = Parameters{
		ExcludeAnnotations: []HasKeyValue{
			{Key: "annotation1", Value: &value2},
		},
	}
	assert.True(t, params.Matches(ns), "Parameters should match the namespace due to excluded annotation")
}

func TestMultipleConditions(t *testing.T) {
	params := Parameters{
		MatchLabels: []HasKeyValue{
			{Key: "key1", Value: nil},
		},
		ExcludeLabels: []HasKeyValue{
			{Key: "key2", Value: &notpresent},
		},
	}
	assert.True(t, params.Matches(ns), "Parameters should match the namespace with multiple conditions")
	params = Parameters{
		MatchLabels: []HasKeyValue{
			{Key: "key1", Value: &value2},
		},
		MatchAnnotations: []HasKeyValue{
			{Key: "annotation1", Value: nil},
			{Key: "annotation2", Value: &value2},
		},
		ExcludeLabels: []HasKeyValue{
			{Key: "key2", Value: &value2},
		},
		ExcludeAnnotations: []HasKeyValue{
			{Key: "annotation2", Value: &notpresent},
		},
	}
	assert.False(t, params.Matches(ns), "Parameters should not match the namespace due to excluded label")
}

func TestEmptyParameters(t *testing.T) {
	params := Parameters{}
	assert.True(t, params.Matches(ns), "Empty parameters should match any namespace")
}
