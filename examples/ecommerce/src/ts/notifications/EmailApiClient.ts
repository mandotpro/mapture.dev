/**
 * @arch.node api email-api
 * @arch.name Email API
 * @arch.domain notifications
 * @arch.owner team-engagement
 * @arch.description Outbound transactional email provider client used by notification-service. It accepts rendered templates and recipient metadata, and provider throttling is the failure mode that needs explicit retry handling.
 */
export class EmailApiClient {
  enqueue(template: string, orderId: string) {
    return `${template}:${orderId}`;
  }
}
