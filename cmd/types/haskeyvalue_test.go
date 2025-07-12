package types

import "testing"

var m = map[string]string{
	"key1": "value",
	"key2": "value",
}

func runHasKeyValueTest(t *testing.T, expected bool, actual bool, failureMessage string) {
	if expected != actual {
		t.Errorf("%s: got %v want %v", failureMessage, actual, expected)
	}
}

func TestHasKeyReturnsTrueWhenKeyIsPresent(t *testing.T) {
	var hkv HasKeyValue = HasKeyValue{
		Key: "key1",
	}

	runHasKeyValueTest(t, true, hkv.HasKey(m), "HasKey returned incorrect result")
}

func TestHasKeyReturnsTrueWhenKeyIsNotPresent(t *testing.T) {
	var hkv HasKeyValue = HasKeyValue{
		Key: "key3",
	}

	runHasKeyValueTest(t, false, hkv.HasKey(m), "HasKey returned incorrect result")
}

func TestHasValueReturnsTrueWhenKeyAndValueArePresent(t *testing.T) {
	val := "value"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "key1",
		Value: &val,
	}

	runHasKeyValueTest(t, true, hkv.HasValue(m), "HasValue returned incorrect result")
}

func TestHasValueReturnsFalseWhenKeyIsPresentAndValueIsNotPresent(t *testing.T) {
	val := "notPresent"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "key1",
		Value: &val,
	}

	runHasKeyValueTest(t, false, hkv.HasValue(m), "HasValue returned incorrect result")
}

func TestHasValueReturnsFalseWhenKeyIsNotPresent(t *testing.T) {
	var hkv HasKeyValue = HasKeyValue{
		Key: "notPresent",
	}

	runHasKeyValueTest(t, false, hkv.HasValue(m), "HasValue returned incorrect result")
}

func TestHasValueReturnsFalseWhenKeyAndValueAreNotPresent(t *testing.T) {
	val := "notPresent"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "notPresent",
		Value: &val,
	}

	runHasKeyValueTest(t, false, hkv.HasValue(m), "HasValue returned incorrect result")
}

func TestMatchesReturnsTrueWhenKeyIsPresentAndValueNotSupplied(t *testing.T) {
	var hkv HasKeyValue = HasKeyValue{
		Key: "key1",
	}

	runHasKeyValueTest(t, true, hkv.Matches(m), "Matches returned incorrect result")
}

func TestMatchesReturnsTrueWhenKeyAndValueArePresent(t *testing.T) {
	val := "value"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "key1",
		Value: &val,
	}

	runHasKeyValueTest(t, true, hkv.Matches(m), "Matches returned incorrect result")
}

func TestMatchesReturnsFalseWhenKeyIsPresentAndValueIsNotPresent(t *testing.T) {
	val := "notPresent"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "key1",
		Value: &val,
	}

	runHasKeyValueTest(t, false, hkv.Matches(m), "Matches returned incorrect result")
}

func TestMatchesReturnsFalseWhenKeyIsNotPresentAndValueIsSupplied(t *testing.T) {
	val := "notPresent"
	var hkv HasKeyValue = HasKeyValue{
		Key:   "notPresent",
		Value: &val,
	}

	runHasKeyValueTest(t, false, hkv.Matches(m), "Matches returned incorrect result")
}
