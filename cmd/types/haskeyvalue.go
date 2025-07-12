package types

type HasKeyValue struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

func (h HasKeyValue) HasKey(m map[string]string) bool {
	_, exists := m[h.Key]
	return exists
}

func (h HasKeyValue) HasValue(m map[string]string) bool {
	if !h.HasKey(m) {
		return false
	}
	if h.Value == nil {
		return false
	}
	value, exists := m[h.Key]
	return exists && value == *h.Value
}

func (h HasKeyValue) Matches(m map[string]string) bool {
	if h.Value == nil {
		return h.HasKey(m) // Key exists, no value to match
	}

	return h.HasValue(m)
}
