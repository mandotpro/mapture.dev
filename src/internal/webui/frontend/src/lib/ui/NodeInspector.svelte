<script lang="ts">
  import type { ImpactPreview, NodeInspectorAction, PresentedNode } from '../types';
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
    actions?: NodeInspectorAction[];
    onaction?: (id: string) => void;
    onclose?: () => void;
  } = $props();

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
</script>

<article class="node-inspector" data-interactive-root>
  <header class="node-inspector__head">
    <div class="node-inspector__title">
      <TokenBadge
        label={badgeLabel}
        icon={typeIcon(node.type)}
        accent={badgeAccent}
        interactive={false}
        compact
        quiet
      />
      <strong>{node.name}</strong>
      <small>{node.id}</small>
      {#if summary}
        <p class="node-inspector__summary">{summary}</p>
      {/if}
    </div>

    <button type="button" class="node-inspector__close" aria-label="Close node overview" onclick={onclose}>
      <span aria-hidden="true">x</span>
    </button>
  </header>

  <div class="node-inspector__grid">
    <section class="node-inspector__meta">
      <span>Domain</span>
      <strong>{domainLabel}</strong>
    </section>
    <section class="node-inspector__meta">
      <span>Owner</span>
      <strong>{ownerLabel}</strong>
    </section>
    <section class="node-inspector__meta">
      <span>{node.kind === 'node' ? 'Source' : 'Composition'}</span>
      <strong>{node.kind === 'node' ? sourceLabel : compositionLabel || 'n/a'}</strong>
    </section>
    <section class="node-inspector__meta">
      <span>Tags</span>
      <strong>{tagLabel || 'n/a'}</strong>
    </section>
  </div>

  {#if impactEnabled}
    <section class="node-inspector__impact">
      <header class="node-inspector__impact-head">
        <div>
          <strong>Impact Preview</strong>
          <p>Quick upstream and downstream reach from the visible graph.</p>
        </div>
        {#if preview.crossBoundaryTouches > 0}
          <TokenBadge
            label="Boundary"
            count={preview.crossBoundaryTouches}
            accent="var(--warning)"
            interactive={false}
            compact
            quiet
          />
        {/if}
      </header>

      <div class="node-inspector__impact-grid">
        <article class="node-inspector__impact-card">
          <span>Downstream</span>
          <strong>{preview.directDownstream.length}</strong>
          <small>{preview.downstreamReach} reachable</small>
        </article>
        <article class="node-inspector__impact-card">
          <span>Upstream</span>
          <strong>{preview.directUpstream.length}</strong>
          <small>{preview.upstreamReach} reachable</small>
        </article>
      </div>

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
              />
            {/each}
          </div>
        </section>
      {/if}

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
              />
            {/each}
          </div>
        </section>
      {/if}
    </section>
  {/if}

  {#if actions.length > 0}
    <div class="node-inspector__actions">
      {#each actions as action}
        <TokenBadge
          label={action.label}
          accent={action.tone === 'accent' ? badgeAccent : 'var(--accent)'}
          interactive
          compact
          trailingText={action.badge ?? ''}
          className="node-inspector__action"
          onclick={() => trigger(action.id)}
        />
      {/each}
    </div>
  {/if}
</article>

<style>
  .node-inspector {
    pointer-events: auto;
    width: min(440px, calc(100vw - 1.5rem));
    display: grid;
    gap: 0.92rem;
    padding: 1rem;
    border-radius: 24px;
    border: 1px solid var(--border-strong);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 7%, transparent), transparent 38%),
      var(--surface-overlay);
    box-shadow: var(--shadow-floating);
    backdrop-filter: blur(18px);
  }

  .node-inspector__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.9rem;
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
    font-size: 0.72rem;
    font-family: "SFMono-Regular", "Consolas", monospace;
    overflow-wrap: anywhere;
  }

  .node-inspector__summary {
    margin: 0.08rem 0 0;
    color: var(--text-secondary);
    font-size: 0.78rem;
    line-height: 1.52;
  }

  .node-inspector__close {
    width: 2.1rem;
    height: 2.1rem;
    padding: 0;
    border-radius: 999px;
    box-shadow: none;
    background: var(--surface-panel);
  }

  .node-inspector__close span {
    font-size: 0.74rem;
    font-weight: 800;
    text-transform: uppercase;
  }

  .node-inspector__grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.42rem;
  }

  .node-inspector__meta {
    display: grid;
    gap: 0.16rem;
    padding: 0.62rem 0.66rem;
    border-radius: 16px;
    border: 1px solid var(--border-soft);
    background: var(--surface-panel-soft);
  }

  .node-inspector__meta span {
    color: var(--text-tertiary);
    font-size: 0.64rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .node-inspector__meta strong {
    color: var(--text-primary);
    font-size: 0.74rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .node-inspector__impact {
    display: grid;
    gap: 0.62rem;
    padding: 0.78rem;
    border-radius: 18px;
    border: 1px solid var(--border-soft);
    background: color-mix(in srgb, var(--surface-panel-soft) 92%, transparent);
  }

  .node-inspector__impact-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.8rem;
  }

  .node-inspector__impact-head strong {
    display: block;
    color: var(--text-primary);
    font-size: 0.78rem;
  }

  .node-inspector__impact-head p {
    margin: 0.14rem 0 0;
    color: var(--text-secondary);
    font-size: 0.7rem;
    line-height: 1.45;
  }

  .node-inspector__impact-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.4rem;
  }

  .node-inspector__impact-card {
    display: grid;
    gap: 0.08rem;
    padding: 0.6rem 0.64rem;
    border-radius: 14px;
    border: 1px solid var(--border-soft);
    background: color-mix(in srgb, var(--surface-raised) 88%, transparent);
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
  }

  @media (max-width: 760px) {
    .node-inspector {
      width: min(100vw - 1rem, 440px);
    }

    .node-inspector__grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .node-inspector__impact-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
