<script lang="ts">
  import type { ImpactPreview, NodeInspectorAction, PresentedNode } from '../types';
  import ActionButton from './ActionButton.svelte';
  import DisclosureButton from './DisclosureButton.svelte';
  import IconButton from './IconButton.svelte';
  import PropertyRow from './PropertyRow.svelte';
  import TokenBadge from './TokenBadge.svelte';

  let {
    node,
    badgeLabel,
    badgeAccent,
    domainLabel,
    ownerLabel,
    sourceLabel,
    tagLabel = '',
    compositionLabel = '',
    summary = '',
    preview,
    impactEnabled = false,
    impactDefaultExpanded = false,
    actions = [],
    onaction,
    onclose,
  }: {
    node: PresentedNode;
    badgeLabel: string;
    badgeAccent: string;
    domainLabel: string;
    ownerLabel: string;
    sourceLabel: string;
    tagLabel?: string;
    compositionLabel?: string;
    summary?: string;
    preview: ImpactPreview;
    impactEnabled?: boolean;
    impactDefaultExpanded?: boolean;
    actions?: NodeInspectorAction[];
    onaction?: (id: string) => void;
    onclose?: () => void;
  } = $props();

  let descriptionExpanded = $state(false);
  let impactExpanded = $state(false);

  function trigger(actionId: string): void {
    onaction?.(actionId);
  }

  function typeIcon(type: string): string {
    const icons: Record<string, string> = {
      service: 'S',
      api: 'A',
      database: 'DB',
      event: 'E',
    };
    return icons[type] ?? 'N';
  }

  function resetExpandedState(): void {
    descriptionExpanded = false;
    impactExpanded = impactEnabled && impactDefaultExpanded;
  }

  $effect(() => {
    node.id;
    impactEnabled;
    impactDefaultExpanded;
    resetExpandedState();
  });

  const hasLongSummary = $derived(summary.trim().length > 140);
  const impactSummary = $derived(buildImpactSummary(preview));

  function buildImpactSummary(current: ImpactPreview): string {
    const parts = [
      `${current.directUpstream.length} upstream`,
      `${current.directDownstream.length} downstream`,
    ];
    if (current.crossBoundaryTouches > 0) {
      parts.push(`${current.crossBoundaryTouches} boundary`);
    }
    return parts.join(' · ');
  }
</script>

<article class="node-inspector" data-interactive-root>
  <header class="node-inspector__head">
    <div class="node-inspector__identity">
      <TokenBadge
        label={badgeLabel}
        icon={typeIcon(node.type)}
        accent={badgeAccent}
        interactive={false}
        compact
        quiet
        className="node-inspector__badge"
      />
      <IconButton className="node-inspector__close" ariaLabel="Close node overview" onclick={onclose} subtle>
        <span class="node-inspector__close-mark" aria-hidden="true">x</span>
      </IconButton>
    </div>

    <div class="node-inspector__title">
      <strong>{node.name}</strong>
      <small>{node.id}</small>
      {#if summary}
        <div class="node-inspector__summary-block">
          <p class={['node-inspector__summary', !descriptionExpanded ? 'is-collapsed' : ''].join(' ')}>
            {summary}
          </p>
          {#if hasLongSummary}
            <button
              type="button"
              class="node-inspector__inline-toggle"
              onclick={() => (descriptionExpanded = !descriptionExpanded)}
            >
              {descriptionExpanded ? 'Show less' : 'Show more'}
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </header>

  <dl class="node-inspector__meta-list">
    <PropertyRow label="Domain" value={domainLabel} />
    <PropertyRow label="Owner" value={ownerLabel} />
    <PropertyRow
      label={node.kind === 'node' ? 'Source' : 'Composition'}
      value={node.kind === 'node' ? sourceLabel : compositionLabel || 'n/a'}
    />
    <PropertyRow label="Tags" value={tagLabel || 'n/a'} />
  </dl>

  {#if impactEnabled}
    <section class="node-inspector__impact">
      <DisclosureButton
        icon="IM"
        title="Impact Preview"
        summary={impactSummary}
        open={impactExpanded}
        className="node-inspector__impact-toggle"
        onclick={() => (impactExpanded = !impactExpanded)}
      />

      {#if impactExpanded}
        <div class="node-inspector__impact-body">
          <div class="node-inspector__impact-grid">
            <article class="node-inspector__impact-card">
              <span>Immediate upstream</span>
              <strong>{preview.directUpstream.length}</strong>
              <small>{preview.upstreamReach} reachable upstream</small>
            </article>
            <article class="node-inspector__impact-card">
              <span>Immediate downstream</span>
              <strong>{preview.directDownstream.length}</strong>
              <small>{preview.downstreamReach} reachable downstream</small>
            </article>
          </div>

          {#if preview.directUpstream.length > 0}
            <section class="node-inspector__impact-list">
              <span>Immediate upstream</span>
              <div class="node-inspector__impact-chips">
                {#each preview.directUpstream.slice(0, 4) as previewNode}
                  <TokenBadge
                    label={previewNode.name}
                    accent={previewNode.colorHint || 'var(--accent)'}
                    interactive={false}
                    compact
                    quiet
                  />
                {/each}
              </div>
            </section>
          {/if}

          {#if preview.directDownstream.length > 0}
            <section class="node-inspector__impact-list">
              <span>Immediate downstream</span>
              <div class="node-inspector__impact-chips">
                {#each preview.directDownstream.slice(0, 4) as previewNode}
                  <TokenBadge
                    label={previewNode.name}
                    accent={previewNode.colorHint || 'var(--accent)'}
                    interactive={false}
                    compact
                    quiet
                  />
                {/each}
              </div>
            </section>
          {/if}
        </div>
      {/if}
    </section>
  {/if}

  {#if actions.length > 0}
    <div class="node-inspector__actions">
      {#each actions as action}
        <ActionButton
          tone={action.tone === 'accent' ? 'soft' : 'ghost'}
          compact
          className="node-inspector__action"
          onclick={() => trigger(action.id)}
        >
          <span>{action.label}</span>
          {#if action.badge}
            <span class="node-inspector__action-badge">{action.badge}</span>
          {/if}
        </ActionButton>
      {/each}
    </div>
  {/if}
</article>

<style>
  .node-inspector {
    pointer-events: auto;
    width: min(440px, calc(100vw - 1.5rem));
    display: grid;
    gap: 0.78rem;
    padding: 0.96rem 1rem;
    border-radius: 24px;
    border: 1px solid var(--border-strong);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 7%, transparent), transparent 38%),
      var(--surface-overlay);
    box-shadow: var(--shadow-floating);
    backdrop-filter: blur(18px);
  }

  .node-inspector__head {
    display: grid;
    gap: 0.7rem;
  }

  .node-inspector__identity {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.9rem;
  }

  .node-inspector__badge {
    justify-self: start;
  }

  .node-inspector__title {
    display: grid;
    gap: 0.24rem;
    min-width: 0;
  }

  .node-inspector__title strong {
    font-family: "Iowan Old Style", "Palatino Linotype", serif;
    font-size: 1.08rem;
    line-height: 1.2;
    color: var(--text-primary);
  }

  .node-inspector__title small {
    color: var(--text-secondary);
    font-size: 0.62rem;
    font-family: "SFMono-Regular", "Consolas", monospace;
    overflow-wrap: anywhere;
    opacity: 0.74;
  }

  .node-inspector__summary-block {
    display: grid;
    justify-items: start;
    gap: 0.1rem;
  }

  .node-inspector__summary {
    margin: 0.08rem 0 0;
    color: var(--text-secondary);
    font-size: 0.78rem;
    line-height: 1.52;
  }

  .node-inspector__summary.is-collapsed {
    display: -webkit-box;
    overflow: hidden;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .node-inspector__inline-toggle {
    min-height: auto;
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    color: var(--accent);
    font-size: 0.72rem;
    font-weight: 700;
    transform: none;
    cursor: pointer;
    transition: color var(--ui-transition-fast);
  }

  .node-inspector__inline-toggle:hover {
    transform: none;
    border-color: transparent;
    box-shadow: none;
    color: var(--accent-strong);
  }

  .node-inspector__close-mark {
    font-size: 0.86rem;
    font-weight: 800;
    text-transform: uppercase;
  }

  .node-inspector__meta-list {
    display: grid;
    gap: 0.52rem;
    margin: 0;
  }

  .node-inspector__impact {
    display: grid;
    gap: 0.52rem;
    padding-top: 0.16rem;
    border-top: 1px solid color-mix(in srgb, var(--border-soft) 82%, transparent);
  }

  .node-inspector__impact-toggle {
    min-width: 0;
  }

  .node-inspector__impact-body {
    display: grid;
    gap: 0.58rem;
  }

  .node-inspector__impact-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.4rem;
  }

  .node-inspector__impact-card {
    display: grid;
    gap: 0.08rem;
    padding: 0.56rem 0.62rem;
    border-radius: 14px;
    border: 1px solid var(--border-soft);
    background: color-mix(in srgb, var(--surface-panel-soft) 82%, transparent);
  }

  .node-inspector__impact-card span {
    color: var(--text-tertiary);
    font-size: 0.64rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .node-inspector__impact-card strong {
    color: var(--text-primary);
    font-size: 0.96rem;
  }

  .node-inspector__impact-card small {
    color: var(--text-secondary);
    font-size: 0.68rem;
  }

  .node-inspector__impact-list {
    display: grid;
    gap: 0.3rem;
  }

  .node-inspector__impact-list > span {
    color: var(--text-tertiary);
    font-size: 0.64rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .node-inspector__impact-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.34rem;
  }

  .node-inspector__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.42rem;
    padding-top: 0.16rem;
    border-top: 1px solid color-mix(in srgb, var(--border-soft) 82%, transparent);
  }

  .node-inspector__action {
    gap: 0.34rem;
  }

  .node-inspector__action-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.2rem;
    height: 1.06rem;
    padding: 0 0.28rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--warning) 14%, var(--surface-panel));
    color: color-mix(in srgb, var(--warning) 86%, var(--text-primary));
    font-size: 0.56rem;
    font-weight: 800;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  @media (max-width: 760px) {
    .node-inspector {
      width: min(100vw - 1rem, 440px);
    }

    .node-inspector__impact-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
