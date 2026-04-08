package graph

import "testing"

func TestEdgeDisplayLabel(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		EdgeCalls:     "calls",
		EdgeDependsOn: "depends on",
		EdgeStoresIn:  "stores in",
		EdgeReadsFrom: "reads from",
		EdgeEmits:     "emits",
		EdgeConsumes:  "consumed by",
		"unknown":     "unknown",
	}

	for edgeType, want := range cases {
		if got := EdgeDisplayLabel(edgeType); got != want {
			t.Fatalf("EdgeDisplayLabel(%q) = %q, want %q", edgeType, got, want)
		}
	}
}
