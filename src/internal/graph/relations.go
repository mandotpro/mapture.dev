package graph

// EdgeDisplayLabel returns the user-facing label for a normalized edge type.
func EdgeDisplayLabel(edgeType string) string {
	switch edgeType {
	case EdgeCalls:
		return "calls"
	case EdgeDependsOn:
		return "depends on"
	case EdgeStoresIn:
		return "stores in"
	case EdgeReadsFrom:
		return "reads from"
	case EdgeEmits:
		return "emits"
	case EdgeConsumes:
		return "consumed by"
	default:
		return edgeType
	}
}
