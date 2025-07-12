package types

type Parameters struct {
	MatchLabels        []HasKeyValue `json:"matchLabels,omitempty"`
	MatchAnnotations   []HasKeyValue `json:"matchAnnotations,omitempty"`
	ExcludeLabels      []HasKeyValue `json:"excludeLabels,omitempty"`
	ExcludeAnnotations []HasKeyValue `json:"excludeAnnotations,omitempty"`
}

func (p Parameters) Matches(ns NamespaceInfo) bool {
	for _, label := range p.MatchLabels {
		if !label.Matches(ns.Labels) {
			return false
		}
	}

	for _, annotation := range p.MatchAnnotations {
		if !annotation.Matches(ns.Annotations) {
			return false
		}
	}

	for _, label := range p.ExcludeLabels {
		if label.Matches(ns.Labels) {
			return false
		}
	}

	for _, annotation := range p.ExcludeAnnotations {
		if annotation.Matches(ns.Annotations) {
			return false
		}
	}

	return true
}
