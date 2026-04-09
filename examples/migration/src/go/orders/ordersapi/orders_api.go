package ordersapi

// @arch.node api orders-api
// @arch.name Orders API
// @arch.domain orders
// @arch.owner team-platform
// @arch.description Modern order write boundary for new client traffic. It accepts direct order creation requests, and during the strangler rollout it can still fall back to legacy-storefront for routes that have not been ported yet.
// @arch.calls service orders-service
// @arch.calls service legacy-storefront
type Handler struct {
	routes map[string]string
}

func (h *Handler) RouteOrder(path string) string {
	return h.routes[path]
}
