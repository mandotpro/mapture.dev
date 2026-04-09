/**
 * This file intentionally demonstrates pub/sub rather than a direct hand-off.
 * The projector is only one subscriber, and notification-service does not know which consumers read notification.sent.
 */
/**
 * @arch.node service notification-audit-projector
 * @arch.name Notification Audit Projector
 * @arch.domain notifications
 * @arch.owner team-engagement
 * @arch.description Builds internal delivery timelines from published notification lifecycle events. Its primary input is notification.sent, and lagging projections are the failure mode support tooling notices first.
 * @arch.depends_on service notification-service
 */
export class NotificationAuditProjector {
  constructor(private readonly timeline: Map<string, string>) {}

  /**
   * @event.id notification.sent
   * @event.role subscriber
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.topic notifications.lifecycle
   * @event.notes This subscriber consumes the published lifecycle event to build a support-facing audit trail without notification-service knowing this projector exists.
   */
  project(eventId: string) { this.timeline.set(eventId, "delivered"); }
}
