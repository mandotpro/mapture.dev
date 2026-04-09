package paymentapi

// @arch.node api payment-api
// @arch.name Payment API
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Thin internal HTTP boundary that accepts checkout payment requests and forwards them to payment-service. Its primary input is an order ready for capture, and the main failure mode is surfacing gateway rejections without duplicating charge attempts.
// @arch.calls service payment-service
type Handler struct {
	service interface {
		CaptureForOrder(string)
	}
}

func (h *Handler) AuthorizeAndCapture(orderID string) {
	h.service.CaptureForOrder(orderID)
}
