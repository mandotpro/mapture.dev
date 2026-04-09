package stripebridge

// @arch.node api stripe-webhook-api
// @arch.name Stripe Webhook API
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Ingress for raw Stripe callbacks before they are normalized for payment-service. Its primary input is an external webhook payload, and the key failure mode is accepting malformed gateway traffic into billing state.
// @arch.calls service payment-service
type WebhookHandler struct {
	forwarder interface {
		Normalize([]byte)
	}
}

// @event.id stripe.webhook.received
// @event.role bridge-in
// @event.domain billing
// @event.owner team-commerce
// @event.topic stripe.webhooks
// @event.notes This bridge-in marks the exact point where external Stripe traffic becomes an internal billing signal that payment-service can validate and reconcile.
func (h *WebhookHandler) HandleStripeWebhook(payload []byte) { h.forwarder.Normalize(payload) }
