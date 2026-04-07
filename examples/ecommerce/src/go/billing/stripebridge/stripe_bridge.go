package stripebridge

// @arch.node api stripe-webhook-api
// @arch.name Stripe Webhook API
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Receives Stripe webhooks and normalizes them before payment-service handles them.
// @arch.calls service payment-service
type WebhookHandler struct{}

// @event.id stripe.webhook.received
// @event.role bridge-in
// @event.domain billing
// @event.owner team-commerce
// @event.topic stripe.webhooks
// @event.notes This is the point where external gateway traffic enters the monorepo.
func (h *WebhookHandler) HandleStripeWebhook(payload []byte) {}
